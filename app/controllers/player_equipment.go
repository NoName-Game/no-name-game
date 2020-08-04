package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

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
func (c *PlayerEquipmentController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.inventory.equip",
		c.Payload,
		[]string{},
		player,
		update,
	) {
		return
	}

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(1, &PlayerController{}) {
			return
		}
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.CurrentState.Payload, &c.Payload)

	// Validate
	var hasError bool
	hasError, err = c.Validator()
	if err != nil {
		panic(err)
	}

	// Se ritornano degli errori
	if hasError {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
		validatorMsg.ReplyMarkup = c.Validation.ReplyKeyboard

		_, err = services.SendMessage(validatorMsg)
		if err != nil {
			panic(err)
		}

		return
	}

	// Ok! Run!
	err = c.Stage()
	if err != nil {
		panic(err)
	}

	// Aggiorno stato finale
	payloadUpdated, _ := json.Marshal(c.Payload)
	c.CurrentState.Payload = string(payloadUpdated)

	rUpdatePlayerState, err := services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
		PlayerState: c.CurrentState,
	})
	if err != nil {
		panic(err)
	}
	c.CurrentState = rUpdatePlayerState.GetPlayerState()

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *PlayerEquipmentController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		),
	)

	switch c.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	// Verifico che la tipologia di equip che vuole il player esista
	case 1:
		if helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "armors"),
			helpers.Trans(c.Player.Language.Slug, "weapons"),
		}) {
			return false, err
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true, err

	// Verifico che il player voglia continuare con l'equip
	case 2:
		if strings.Contains(c.Update.Message.Text, "🩸") || strings.Contains(c.Update.Message.Text, "🛡") {
			return false, err
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true, err

	// Verifico la conferma dell'equip
	case 3:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "confirm") {
			return false, err
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true, err
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *PlayerEquipmentController) Stage() (err error) {
	switch c.CurrentState.Stage {
	// In questo stage faccio un micro recap al player del suo equipaggiamento
	// attuale e mostro a tastierino quale categoria vorrebbe equipaggiare
	case 0:
		// ******************
		// Recupero armatura equipaggiata
		// ******************
		var currentArmorsEquipment string
		currentArmorsEquipment = fmt.Sprintf("%s:\n", helpers.Trans(c.Player.Language.Slug, "armor"))

		rGetPlayerArmors, err := services.NnSDK.GetPlayerArmors(helpers.NewContext(1), &pb.GetPlayerArmorsRequest{
			PlayerID: c.Player.GetID(),
			Equipped: true,
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

		rGetPlayerWeapons, err := services.NnSDK.GetPlayerWeapons(helpers.NewContext(1), &pb.GetPlayerWeaponsRequest{
			PlayerID: c.Player.GetID(),
			Equipped: true,
		})
		if err != nil {
			return err
		}

		if len(rGetPlayerWeapons.GetWeapons()) > 0 {
			for _, weapon := range rGetPlayerWeapons.GetWeapons() {
				currentWeaponsEquipment += fmt.Sprintf(
					"- %s (*%s*) %v🩸 \n",
					weapon.Name,
					strings.ToUpper(weapon.Rarity.Slug),
					weapon.RawDamage,
				)
			}
		} else {
			currentWeaponsEquipment += helpers.Trans(c.Player.Language.Slug, "inventory.weapons.zero_equipment")
		}

		// Invio messagio con recap e con selettore categoria
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s \n\n %s",
				helpers.Trans(c.Player.Language.Slug, "inventory.type"),
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
		c.CurrentState.Stage = 1

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
			rGetPlayerArmors, err := services.NnSDK.GetPlayerArmors(helpers.NewContext(1), &pb.GetPlayerArmorsRequest{
				PlayerID: c.Player.GetID(),
				Equipped: false,
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
			rGetPlayerWeapons, err := services.NnSDK.GetPlayerWeapons(helpers.NewContext(1), &pb.GetPlayerWeaponsRequest{
				PlayerID: c.Player.GetID(),
				Equipped: false,
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
		c.CurrentState.Stage = 2

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
			rFindArmorByName, err := services.NnSDK.FindArmorByName(helpers.NewContext(1), &pb.FindArmorByNameRequest{
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
			rFindWeaponByName, err := services.NnSDK.FindWeaponByName(helpers.NewContext(1), &pb.FindWeaponByNameRequest{
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
		c.CurrentState.Stage = 3

	// In questo stage se l'utente ha confermato continuo con l'equipaggiamento
	// TODO: bisogna verifica che ci sia solo 1 arma o armatura equipaggiata
	case 3:
		switch c.Payload.Type {
		case helpers.Trans(c.Player.Language.Slug, "armors"):
			_, err := services.NnSDK.GetArmorByID(helpers.NewContext(1), &pb.GetArmorByIDRequest{
				ID: c.Payload.EquipID,
			})
			if err != nil {
				log.Fatalf("could not greet: %v", err)
			}

			// Aggiorno equipped
			_, err = services.NnSDK.UpdateArmor(helpers.NewContext(1), &pb.UpdateArmorRequest{
				Armor: &pb.Armor{Equipped: true},
			})
			if err != nil {
				return err
			}
		case helpers.Trans(c.Player.Language.Slug, "weapons"):
			_, err := services.NnSDK.GetWeaponByID(helpers.NewContext(1), &pb.GetWeaponByIDRequest{
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
		c.CurrentState.Completed = true
	}

	return
}
