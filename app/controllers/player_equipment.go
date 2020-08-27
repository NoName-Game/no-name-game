package controllers

import (
	"fmt"
	"strings"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerEquipmentController
// ====================================
type PlayerEquipmentController struct {
	BaseController
	Payload struct {
		Type    string
		EquipID uint32
	}
}

// ====================================
// Handle
// ====================================
func (c *PlayerEquipmentController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller: "route.inventory.equip",
		ControllerBack: ControllerBack{
			To:        &PlayerController{},
			FromStage: 1,
		},
		Payload: c.Payload,
	}) {
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)

	// Validate
	var hasError bool
	if hasError = c.Validator(); hasError {
		c.Validate()
		return
	}

	// Ok! Run!
	if err = c.Stage(); err != nil {
		panic(err)
	}

	// Completo progressione
	if err = c.Completing(c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *PlayerEquipmentController) Validator() (hasErrors bool) {
	switch c.PlayerData.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	// Verifico che la tipologia di equip che vuole il player esista
	case 1:
		if helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "armors"),
			helpers.Trans(c.Player.Language.Slug, "weapons"),
		}) {
			return false
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true

	// Verifico che il player voglia continuare con l'equip
	case 2:
		if strings.Contains(c.Update.Message.Text, "🩸") || strings.Contains(c.Update.Message.Text, "🛡") {
			return false
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true

	// Verifico la conferma dell'equip
	case 3:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "confirm") {
			return false
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *PlayerEquipmentController) Stage() (err error) {
	switch c.PlayerData.CurrentState.Stage {
	// In questo stage faccio un micro recap al player del suo equipaggiamento
	// attuale e mostro a tastierino quale categoria vorrebbe equipaggiare
	case 0:
		// ******************
		// Recupero armatura equipaggiata
		// ******************
		var currentArmorsEquipment string
		currentArmorsEquipment = fmt.Sprintf("%s:\n", helpers.Trans(c.Player.Language.Slug, "armor"))

		var rGetPlayerArmors *pb.GetPlayerArmorsEquippedResponse
		rGetPlayerArmors, err = services.NnSDK.GetPlayerArmorsEquipped(helpers.NewContext(1), &pb.GetPlayerArmorsEquippedRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		// armatura base player
		if len(rGetPlayerArmors.GetArmors()) > 0 {
			var helmet, gauntlets, chest, leg string
			for _, armor := range rGetPlayerArmors.GetArmors() {
				switch armor.ArmorCategory.Slug {
				case "helmet":
					helmet = fmt.Sprintf("%s \nDef: *%v* | Evs: *%v* | Hlv: *%v* | Pot: *%v* ", armor.Name, armor.Defense, armor.Evasion, armor.Halving, armor.Potential)
				case "gauntlets":
					gauntlets = fmt.Sprintf("%s \nDef: *%v* | Evs: *%v* | Hlv: *%v* | Pot: *%v* ", armor.Name, armor.Defense, armor.Evasion, armor.Halving, armor.Potential)
				case "chest":
					chest = fmt.Sprintf("%s \nDef: *%v* | Evs: *%v* | Hlv: *%v* | Pot: *%v* ", armor.Name, armor.Defense, armor.Evasion, armor.Halving, armor.Potential)
				case "leg":
					leg = fmt.Sprintf("%s \nDef: *%v* | Evs: *%v* | Hlv: *%v* | Pot: *%v* ", armor.Name, armor.Defense, armor.Evasion, armor.Halving, armor.Potential)
				}
			}

			currentArmorsEquipment += fmt.Sprintf("%s \n\n%s \n\n%s \n\n%s",
				helmet,
				gauntlets,
				chest,
				leg,
			)
		} else {
			currentArmorsEquipment += helpers.Trans(c.Player.Language.Slug, "inventory.armors.zero_equipment")
		}

		// ******************
		// Recupero armi equipaggiate
		// ******************
		var currentWeaponsEquipment string
		currentWeaponsEquipment = fmt.Sprintf("%s:\n", helpers.Trans(c.Player.Language.Slug, "weapon"))

		var rGetPlayerWeaponEquippedResponse *pb.GetPlayerWeaponEquippedResponse
		rGetPlayerWeaponEquippedResponse, err = services.NnSDK.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		if rGetPlayerWeaponEquippedResponse.GetWeapon() != nil {
			currentWeaponsEquipment += fmt.Sprintf(
				"- %s (*%s*) %v🩸 \n",
				rGetPlayerWeaponEquippedResponse.GetWeapon().GetName(),
				strings.ToUpper(rGetPlayerWeaponEquippedResponse.GetWeapon().GetRarity().GetSlug()),
				rGetPlayerWeaponEquippedResponse.GetWeapon().GetRawDamage(),
			)
		} else {
			currentWeaponsEquipment += helpers.Trans(c.Player.Language.Slug, "inventory.weapons.zero_equipment")
		}

		// Statistica totale armatura
		var defense, evasion, halving float32
		if len(rGetPlayerArmors.GetArmors()) > 0 {
			for _, armor := range rGetPlayerArmors.GetArmors() {
				defense += armor.Defense
				evasion += armor.Evasion
				halving += armor.Halving
			}
		}

		// Invio messagio con recap e con selettore categoria
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s \n\n%s \n %s",
				helpers.Trans(c.Player.Language.Slug, "inventory.type"),
				fmt.Sprintf("🛡 Def: *%v* | Evs: *%v* | Hlv: *%v*\n",
					defense, evasion, halving,
				),
				fmt.Sprintf(
					"%s\n%s\n\n%s",
					helpers.Trans(c.Player.Language.Slug, "inventory.equip.equipped"),
					currentArmorsEquipment,
					currentWeaponsEquipment,
				),
			),
		)
		msg.ParseMode = "markdown"

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "armors")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "weapons")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Avanzo di stage
		c.PlayerData.CurrentState.Stage = 1

	// In questo stage chiedo di indicarmi quale armatura o arma intende equipaggiare
	case 1:
		var mainMessage string
		// Costruisco keyboard risposta
		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		c.Payload.Type = c.Update.Message.Text

		switch c.Payload.Type {
		case helpers.Trans(c.Player.Language.Slug, "armors"):
			mainMessage = helpers.Trans(c.Player.Language.Slug, "inventory.armors.no_one")

			// Recupero nuovamente armature player, richiamando la rotta dedicata
			// in questa maniera posso filtrare per quelle che non sono equipaggiate
			var rGetPlayerArmors *pb.GetPlayerArmorsResponse
			rGetPlayerArmors, err = services.NnSDK.GetPlayerArmors(helpers.NewContext(1), &pb.GetPlayerArmorsRequest{
				PlayerID: c.Player.GetID(),
			})
			if err != nil {
				return err
			}

			if len(rGetPlayerArmors.GetArmors()) > 0 {
				mainMessage = helpers.Trans(c.Player.Language.Slug, "inventory.armors.what")

				// Ciclo armature del player
				for _, armor := range rGetPlayerArmors.GetArmors() {
					keyboardRow := tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							fmt.Sprintf(
								"%s (%s) 🛡",
								// helpers.Trans(c.Player.Language.Slug, "equip"),
								armor.Name,
								strings.ToUpper(armor.Rarity.Slug),
							),
						),
					)
					keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
				}
			}

		case helpers.Trans(c.Player.Language.Slug, "weapons"):
			mainMessage = helpers.Trans(c.Player.Language.Slug, "inventory.armors.no_one")

			// Recupero nuovamente armi player, richiamando la rotta dedicata
			// in questa maniera posso filtrare per quelle che non sono equipaggiate
			var rGetPlayerWeapons *pb.GetPlayerWeaponsResponse
			rGetPlayerWeapons, err = services.NnSDK.GetPlayerWeapons(helpers.NewContext(1), &pb.GetPlayerWeaponsRequest{
				PlayerID: c.Player.GetID(),
			})
			if err != nil {
				return err
			}

			if len(rGetPlayerWeapons.GetWeapons()) > 0 {
				mainMessage = helpers.Trans(c.Player.Language.Slug, "inventory.weapons.what")

				// Ciclo armi player
				for _, weapon := range rGetPlayerWeapons.GetWeapons() {
					keyboardRow := tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							fmt.Sprintf(
								"%s (%s) %v🩸",
								// helpers.Trans(c.Player.Language.Slug, "equip"),
								weapon.Name,
								strings.ToUpper(weapon.Rarity.Slug),
								weapon.RawDamage,
							),
						),
					)
					keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
				}
			}
		}

		// Aggiungo tasti back and clears
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, mainMessage)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.PlayerData.CurrentState.Stage = 2

	// In questo stato ricerco effettivamente l'arma o l'armatura che il player vuole
	// equipaggiare e me lo metto nel payload in attesa di conferma
	case 2:
		var equipmentName string
		var equipmentID uint32
		var equipmentError bool

		// Ripulisco messaggio per recupermi solo il nome
		equipmentName = strings.Split(c.Update.Message.Text, " (")[0]

		switch c.Payload.Type {
		case helpers.Trans(c.Player.Language.Slug, "armors"):
			var rFindArmorByName *pb.FindArmorByNameResponse
			rFindArmorByName, err = services.NnSDK.FindArmorByName(helpers.NewContext(1), &pb.FindArmorByNameRequest{
				Name: equipmentName,
			})
			if err != nil {
				return err
			}

			// Verifico se appartiene correttamente al player
			if rFindArmorByName.GetArmor().GetPlayerID() != c.Player.ID {
				equipmentError = true
			}

			equipmentID = rFindArmorByName.GetArmor().GetID()
		case helpers.Trans(c.Player.Language.Slug, "weapons"):
			var rFindWeaponByName *pb.FindWeaponByNameResponse
			rFindWeaponByName, err = services.NnSDK.FindWeaponByName(helpers.NewContext(1), &pb.FindWeaponByNameRequest{
				Name: equipmentName,
			})
			if err != nil {
				return err
			}

			// Verifico se appartiene correttamente al player
			if rFindWeaponByName.GetWeapon().GetPlayerID() != c.Player.ID {
				equipmentError = true
			}

			equipmentID = rFindWeaponByName.GetWeapon().GetID()
		}

		if equipmentError {
			// Invio messaggio error
			msg := services.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "inventory.equip.error"),
			)
			msg.ParseMode = "markdown"

			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
				),
			)

			_, err = services.SendMessage(msg)
			if err != nil {
				return err
			}

			return
		}

		// Invio messaggio per conferma equipaggiamento
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "inventory.equip.confirm", equipmentName),
		)
		msg.ParseMode = "markdown"

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.Payload.EquipID = equipmentID
		c.PlayerData.CurrentState.Stage = 3

	// In questo stage se l'utente ha confermato continuo con l'equipaggiamento
	// TODO: bisogna verifica che ci sia solo 1 arma o armatura equipaggiata
	case 3:
		switch c.Payload.Type {
		case helpers.Trans(c.Player.Language.Slug, "armors"):
			_, err = services.NnSDK.GetArmorByID(helpers.NewContext(1), &pb.GetArmorByIDRequest{
				ID: c.Payload.EquipID,
			})
			if err != nil {
				return err
			}

			// Aggiorno equipped
			_, err = services.NnSDK.UpdateArmor(helpers.NewContext(1), &pb.UpdateArmorRequest{
				Armor: &pb.Armor{Equipped: true},
			})
			if err != nil {
				return err
			}
		case helpers.Trans(c.Player.Language.Slug, "weapons"):
			_, err = services.NnSDK.GetWeaponByID(helpers.NewContext(1), &pb.GetWeaponByIDRequest{
				ID: c.Payload.EquipID,
			})
			if err != nil {
				return err
			}

			// Aggiorno equipped
			_, err = services.NnSDK.UpdateWeapon(helpers.NewContext(1), &pb.UpdateWeaponRequest{
				Weapon: &pb.Weapon{Equipped: true},
			})
			if err != nil {
				return err
			}
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "inventory.equip.completed"),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.PlayerData.CurrentState.Completed = true

		// ###################
		// TUTORIAL - Solo il player si trova dentro il tutorial forzo di tornarare al menu
		// ###################
		if c.InTutorial() {
			c.Configuration.ControllerBack.To = &MenuController{}
		}
	}

	return
}
