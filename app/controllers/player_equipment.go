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
		ItemType     string // Armor/Weapon
		ItemCategory string // Head/Leg/Chest/Arms
		EquipID      uint32
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

	// Carico payload
	if err = helpers.GetPayloadController(c.Player.ID, c.CurrentState.Controller, &c.Payload); err != nil {
		panic(err)
	}

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
	if err = c.Completing(&c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *PlayerEquipmentController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false
	case 1:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "armors") {
			c.Payload.ItemType = "armors"
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "weapons") {
			// Se viene richiesto di craftare un'arma passo direttamente alla lista delle armi
			// in quanto le armi non hanno una categoria
			c.Payload.ItemType = "weapons"
			c.CurrentState.Stage = 2
			return false
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true
	// Verifico che la tipologia di equip che vuole il player esista
	case 2:
		if c.Payload.ItemCategory = helpers.CheckAndReturnCategorySlug(c.Player.Language.Slug, c.Update.Message.Text); c.Payload.ItemCategory != "" {
			return false
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true
	// Verifico che il player voglia continuare con l'equip
	case 3:
		if strings.Contains(c.Update.Message.Text, "🩸") || strings.Contains(c.Update.Message.Text, "🛡") {
			return false
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true

	// Verifico la conferma dell'equip
	case 4:
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
	switch c.CurrentState.Stage {
	// In questo stage faccio un micro recap al player del suo equipaggiamento
	// attuale e mostro a tastierino quale categoria vorrebbe equipaggiare
	case 0:
		// ******************
		// Recupero armatura equipaggiata
		// ******************
		var currentArmorsEquipment string
		currentArmorsEquipment = fmt.Sprintf("*%s*:", helpers.Trans(c.Player.Language.Slug, "armor"))

		var rGetPlayerArmors *pb.GetPlayerArmorsEquippedResponse
		if rGetPlayerArmors, err = services.NnSDK.GetPlayerArmorsEquipped(helpers.NewContext(1), &pb.GetPlayerArmorsEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			return err
		}

		// armatura base player
		if len(rGetPlayerArmors.GetArmors()) > 0 {
			// var head, gauntlets, chest, leg string
			for _, armor := range rGetPlayerArmors.GetArmors() {
				currentArmorsEquipment += fmt.Sprintf("\n%s\n*%s* (*%s*)\nDEF: *%.2v* | EVS: *%.2v* | HLV: *%.2v*\n",
					helpers.Trans(c.Player.Language.Slug, armor.GetArmorCategory().GetSlug()),
					armor.Name, strings.ToUpper(armor.GetRarity().GetSlug()),
					armor.Defense,
					armor.Evasion,
					armor.Halving,
				)
			}
		} else {
			currentArmorsEquipment += helpers.Trans(c.Player.Language.Slug, "inventory.armors.zero_equipment")
		}

		// ******************
		// Recupero arma equipaggiata
		// ******************
		var currentWeaponsEquipment string
		currentWeaponsEquipment = fmt.Sprintf("*%s*:", helpers.Trans(c.Player.Language.Slug, "weapon"))

		var rGetPlayerWeaponEquippedResponse *pb.GetPlayerWeaponEquippedResponse
		if rGetPlayerWeaponEquippedResponse, err = services.NnSDK.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			return err
		}

		if rGetPlayerWeaponEquippedResponse.GetWeapon() != nil {
			currentWeaponsEquipment += fmt.Sprintf(
				"\n*%s* (*%s*)\nDMG:*%.2v* | PCS: *%.2v* | PNT: *%.2v*",
				rGetPlayerWeaponEquippedResponse.GetWeapon().GetName(),
				strings.ToUpper(rGetPlayerWeaponEquippedResponse.GetWeapon().GetRarity().GetSlug()),
				rGetPlayerWeaponEquippedResponse.GetWeapon().GetRawDamage(),
				rGetPlayerWeaponEquippedResponse.GetWeapon().GetPrecision(),
				rGetPlayerWeaponEquippedResponse.GetWeapon().GetPenetration(),
			)
		} else {
			currentWeaponsEquipment += helpers.Trans(c.Player.Language.Slug, "inventory.weapons.zero_equipment")
		}

		// Invio messagio con recap e con selettore categoria
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s\n\n%s\n%s",
				helpers.Trans(c.Player.Language.Slug, "inventory.equip.equipped"),
				currentArmorsEquipment,
				currentWeaponsEquipment,
			),
		)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "armors")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "weapons")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1

	// In questo stage chiedo di indicarmi quale armatura o arma intende equipaggiare
	case 1:
		var message string
		var keyboardRowCategories [][]tgbotapi.KeyboardButton

		switch c.Payload.ItemType {
		case "armors":
			message = helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.armor.type")

			var rGetAllArmorCategory *pb.GetAllArmorCategoryResponse
			if rGetAllArmorCategory, err = services.NnSDK.GetAllArmorCategory(helpers.NewContext(1), &pb.GetAllArmorCategoryRequest{}); err != nil {
				return
			}

			for _, category := range rGetAllArmorCategory.GetArmorCategories() {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, category.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		msg := services.NewMessage(c.Player.ChatID, message)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		if _, err = services.SendMessage(msg); err != nil {
			return
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		var mainMessage string
		// Costruisco keyboard risposta
		var keyboardRowCategories [][]tgbotapi.KeyboardButton

		switch c.Payload.ItemType {
		// **************************
		// GESTIONE ARMATURE
		// **************************
		case "armors":
			mainMessage = helpers.Trans(c.Player.Language.Slug, "inventory.armors.no_one")

			// Recupero categoria scelta dal player
			var rGetArmorCategoryBySlugRequest *pb.GetArmorCategoryBySlugResponse
			if rGetArmorCategoryBySlugRequest, err = services.NnSDK.GetArmorCategoryBySlug(helpers.NewContext(1), &pb.GetArmorCategoryBySlugRequest{
				Slug: c.Payload.ItemCategory,
			}); err != nil {
				return err
			}

			// Recupero armature non equipaggiate filtrate per la categoria scelta
			var rGetPlayerArmorsByCategoryID *pb.GetPlayerArmorsByCategoryIDResponse
			if rGetPlayerArmorsByCategoryID, err = services.NnSDK.GetPlayerArmorsByCategoryID(helpers.NewContext(1), &pb.GetPlayerArmorsByCategoryIDRequest{
				PlayerID:   c.Player.GetID(),
				CategoryID: rGetArmorCategoryBySlugRequest.GetArmorCategory().GetID(),
			}); err != nil {
				return err
			}

			if len(rGetPlayerArmorsByCategoryID.GetArmors()) > 0 {
				mainMessage = helpers.Trans(c.Player.Language.Slug, "inventory.armors.what")

				// Ciclo armature del player
				for _, armor := range rGetPlayerArmorsByCategoryID.GetArmors() {
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

		// **************************
		// GESTIONE ARMI
		// **************************
		case "weapons":
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
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3

	// In questo stato ricerco effettivamente l'arma o l'armatura che il player vuole
	// equipaggiare e me lo metto nel payload in attesa di conferma
	case 3:
		var equipmentName string
		var equipmentID uint32
		var equipmentError bool
		var confirmMessage string

		// Ripulisco messaggio per recupermi solo il nome
		equipmentName = strings.Split(c.Update.Message.Text, " (")[0]
		switch c.Payload.ItemType {
		case "armors":
			var rGetArmorByName *pb.GetArmorByNameResponse
			if rGetArmorByName, err = services.NnSDK.GetArmorByName(helpers.NewContext(1), &pb.GetArmorByNameRequest{
				Name: equipmentName,
			}); err != nil {
				return err
			}

			// Verifico se appartiene correttamente al player
			if rGetArmorByName.GetArmor().GetPlayerID() != c.Player.ID {
				equipmentError = true
			}
			equipmentID = rGetArmorByName.GetArmor().GetID()

			// Recupero armatura attualmente equipaggiata per la categoria scelta
			var rGetPlayerArmorEquippedByCategoryID *pb.GetPlayerArmorEquippedByCategoryIDResponse
			if rGetPlayerArmorEquippedByCategoryID, err = services.NnSDK.GetPlayerArmorEquippedByCategoryID(helpers.NewContext(1), &pb.GetPlayerArmorEquippedByCategoryIDRequest{
				PlayerID:   c.Player.GetID(),
				CategoryID: rGetArmorByName.GetArmor().GetArmorCategory().GetID(),
			}); err != nil {
				return err
			}

			// Preparo messaggio di conferma
			confirmMessage = helpers.Trans(c.Player.Language.Slug, "inventory.equip.confirm", rGetArmorByName.GetArmor().GetName())
			if rGetPlayerArmorEquippedByCategoryID.GetArmor().GetID() > 0 {
				confirmMessage = helpers.Trans(c.Player.Language.Slug, "inventory.equip.confirm_armor",
					rGetArmorByName.GetArmor().GetName(),
					rGetPlayerArmorEquippedByCategoryID.GetArmor().GetName(),
				)
			}
		case "weapons":
			var rGetWeaponByName *pb.GetWeaponByNameResponse
			if rGetWeaponByName, err = services.NnSDK.GetWeaponByName(helpers.NewContext(1), &pb.GetWeaponByNameRequest{
				Name: equipmentName,
			}); err != nil {
				return err
			}

			// Verifico se appartiene correttamente al player
			if rGetWeaponByName.GetWeapon().GetPlayerID() != c.Player.ID {
				equipmentError = true
			}

			equipmentID = rGetWeaponByName.GetWeapon().GetID()

			// Recupero arma attualmente equipaggiata
			var rGetPlayerWeaponEquipped *pb.GetPlayerWeaponEquippedResponse
			if rGetPlayerWeaponEquipped, err = services.NnSDK.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				return err
			}

			// Preparo messaggio di conferma
			confirmMessage = helpers.Trans(c.Player.Language.Slug, "inventory.equip.confirm", rGetWeaponByName.GetWeapon().GetName())
			if rGetPlayerWeaponEquipped.GetWeapon().GetID() > 0 {
				confirmMessage = helpers.Trans(c.Player.Language.Slug, "inventory.equip.confirm_weapon",
					rGetWeaponByName.GetWeapon().GetName(),
					rGetPlayerWeaponEquipped.GetWeapon().GetName(),
				)
			}

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

			if _, err = services.SendMessage(msg); err != nil {
				return
			}

			return
		}

		// Invio messaggio per conferma equipaggiamento
		msg := services.NewMessage(c.Update.Message.Chat.ID, confirmMessage)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		if _, err = services.SendMessage(msg); err != nil {
			return
		}

		// Aggiorno stato
		c.Payload.EquipID = equipmentID
		c.CurrentState.Stage = 4

	// In questo stage se l'utente ha confermato continuo con l'equipaggiamento
	case 4:
		switch c.Payload.ItemType {
		case "armors":
			// Recupero categoria scelta dal player
			var rGetArmorCategoryBySlugRequest *pb.GetArmorCategoryBySlugResponse
			if rGetArmorCategoryBySlugRequest, err = services.NnSDK.GetArmorCategoryBySlug(helpers.NewContext(1), &pb.GetArmorCategoryBySlugRequest{
				Slug: c.Payload.ItemCategory,
			}); err != nil {
				return
			}

			// Recupero armatura attualmente equipaggiata per la categoria scelta
			var rGetPlayerArmorEquippedByCategoryID *pb.GetPlayerArmorEquippedByCategoryIDResponse
			if rGetPlayerArmorEquippedByCategoryID, err = services.NnSDK.GetPlayerArmorEquippedByCategoryID(helpers.NewContext(1), &pb.GetPlayerArmorEquippedByCategoryIDRequest{
				PlayerID:   c.Player.GetID(),
				CategoryID: rGetArmorCategoryBySlugRequest.GetArmorCategory().GetID(),
			}); err != nil {
				return
			}

			// Rimuovo equipaggiamento attuale
			if rGetPlayerArmorEquippedByCategoryID.GetArmor().GetID() > 0 {
				if _, err = services.NnSDK.UpdateArmor(helpers.NewContext(1), &pb.UpdateArmorRequest{
					Armor: &pb.Armor{
						ID:       rGetPlayerArmorEquippedByCategoryID.GetArmor().GetID(),
						Equipped: false,
					},
				}); err != nil {
					return
				}
			}

			// Aggiorno con quello nuovo
			if _, err = services.NnSDK.UpdateArmor(helpers.NewContext(1), &pb.UpdateArmorRequest{
				Armor: &pb.Armor{
					ID:       c.Payload.EquipID,
					Equipped: true,
				},
			}); err != nil {
				return
			}
		case "weapons":
			// Recupero arma attualmente equipaggiata
			var rGetPlayerWeaponEquipped *pb.GetPlayerWeaponEquippedResponse
			if rGetPlayerWeaponEquipped, err = services.NnSDK.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				return
			}

			// Rimovo arma attualmente equipaggiata
			if rGetPlayerWeaponEquipped.GetWeapon().GetID() > 0 {
				if _, err = services.NnSDK.UpdateWeapon(helpers.NewContext(1), &pb.UpdateWeaponRequest{
					Weapon: &pb.Weapon{
						ID:       rGetPlayerWeaponEquipped.GetWeapon().GetID(),
						Equipped: false,
					},
				}); err != nil {
					return
				}
			}

			// Aggiorno equipped
			if _, err = services.NnSDK.UpdateWeapon(helpers.NewContext(1), &pb.UpdateWeaponRequest{
				Weapon: &pb.Weapon{
					ID:       c.Payload.EquipID,
					Equipped: true,
				},
			}); err != nil {
				return
			}
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "inventory.equip.completed"),
		)
		msg.ParseMode = "markdown"

		if _, err = services.SendMessage(msg); err != nil {
			return
		}

		// Completo lo stato
		c.CurrentState.Completed = true

		// ###################
		// TUTORIAL - Solo il player si trova dentro il tutorial forzo di tornarare al menu
		// ###################
		if c.InTutorial() {
			c.Configuration.ControllerBack.To = &MenuController{}
		}
	}

	return
}
