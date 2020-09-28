package controllers

import (
	"fmt"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerEquipmentController
// ====================================
type PlayerEquipmentController struct {
	Controller
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
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.inventory.equip",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerController{},
				FromStage: 1,
			},
		},
	}) {
		return
	}

	// Validate
	if c.Validator() {
		c.Validate()
		return
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *PlayerEquipmentController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico quale tipologia di equipaggiamento si vuole gestire
	// ##################################################################################################
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

		return true
	// ##################################################################################################
	// Verifico che la tipologia di equip che vuole il player esista
	// ##################################################################################################
	case 2:
		if c.Payload.ItemCategory = helpers.CheckAndReturnCategorySlug(c.Player.Language.Slug, c.Update.Message.Text); c.Payload.ItemCategory == "" {
			return true
		}
	// ##################################################################################################
	// Verifico conferma equipaggiamento player
	// ##################################################################################################
	case 3:
		if strings.Contains(c.Update.Message.Text, "🩸") || strings.Contains(c.Update.Message.Text, "🛡") {
			return false
		}

		return true
	// ##################################################################################################
	// Verifico conferma equipaggiamento player
	// ##################################################################################################
	case 4:
		// Verifico la conferma dell'equip
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *PlayerEquipmentController) Stage() {
	var err error
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
		if rGetPlayerArmors, err = config.App.Server.Connection.GetPlayerArmorsEquipped(helpers.NewContext(1), &pb.GetPlayerArmorsEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
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
		if rGetPlayerWeaponEquippedResponse, err = config.App.Server.Connection.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
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
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
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
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
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
			if rGetAllArmorCategory, err = config.App.Server.Connection.GetAllArmorCategory(helpers.NewContext(1), &pb.GetAllArmorCategoryRequest{}); err != nil {
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

		msg := helpers.NewMessage(c.Player.ChatID, message)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
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
			if rGetArmorCategoryBySlugRequest, err = config.App.Server.Connection.GetArmorCategoryBySlug(helpers.NewContext(1), &pb.GetArmorCategoryBySlugRequest{
				Slug: c.Payload.ItemCategory,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Recupero armature non equipaggiate filtrate per la categoria scelta
			var rGetPlayerArmorsByCategoryID *pb.GetPlayerArmorsByCategoryIDResponse
			if rGetPlayerArmorsByCategoryID, err = config.App.Server.Connection.GetPlayerArmorsByCategoryID(helpers.NewContext(1), &pb.GetPlayerArmorsByCategoryIDRequest{
				PlayerID:   c.Player.GetID(),
				CategoryID: rGetArmorCategoryBySlugRequest.GetArmorCategory().GetID(),
			}); err != nil {
				c.Logger.Panic(err)
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
			if rGetPlayerWeapons, err = config.App.Server.Connection.GetPlayerWeapons(helpers.NewContext(1), &pb.GetPlayerWeaponsRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
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
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, mainMessage)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
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
			if rGetArmorByName, err = config.App.Server.Connection.GetArmorByName(helpers.NewContext(1), &pb.GetArmorByNameRequest{
				Name: equipmentName,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Verifico se appartiene correttamente al player
			if rGetArmorByName.GetArmor().GetPlayerID() != c.Player.ID {
				equipmentError = true
			}
			equipmentID = rGetArmorByName.GetArmor().GetID()

			// Recupero armatura attualmente equipaggiata per la categoria scelta
			var rGetPlayerArmorEquippedByCategoryID *pb.GetPlayerArmorEquippedByCategoryIDResponse
			if rGetPlayerArmorEquippedByCategoryID, err = config.App.Server.Connection.GetPlayerArmorEquippedByCategoryID(helpers.NewContext(1), &pb.GetPlayerArmorEquippedByCategoryIDRequest{
				PlayerID:   c.Player.GetID(),
				CategoryID: rGetArmorByName.GetArmor().GetArmorCategory().GetID(),
			}); err != nil {
				c.Logger.Panic(err)
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
			if rGetWeaponByName, err = config.App.Server.Connection.GetWeaponByName(helpers.NewContext(1), &pb.GetWeaponByNameRequest{
				Name: equipmentName,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Verifico se appartiene correttamente al player
			if rGetWeaponByName.GetWeapon().GetPlayerID() != c.Player.ID {
				equipmentError = true
			}

			equipmentID = rGetWeaponByName.GetWeapon().GetID()

			// Recupero arma attualmente equipaggiata
			var rGetPlayerWeaponEquipped *pb.GetPlayerWeaponEquippedResponse
			if rGetPlayerWeaponEquipped, err = config.App.Server.Connection.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
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
			msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "inventory.equip.error"))
			msg.ParseMode = "markdown"
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
				),
			)

			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}

			return
		}

		// Invio messaggio per conferma equipaggiamento
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, confirmMessage)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
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
			if rGetArmorCategoryBySlugRequest, err = config.App.Server.Connection.GetArmorCategoryBySlug(helpers.NewContext(1), &pb.GetArmorCategoryBySlugRequest{
				Slug: c.Payload.ItemCategory,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Recupero armatura attualmente equipaggiata per la categoria scelta
			var rGetPlayerArmorEquippedByCategoryID *pb.GetPlayerArmorEquippedByCategoryIDResponse
			if rGetPlayerArmorEquippedByCategoryID, err = config.App.Server.Connection.GetPlayerArmorEquippedByCategoryID(helpers.NewContext(1), &pb.GetPlayerArmorEquippedByCategoryIDRequest{
				PlayerID:   c.Player.GetID(),
				CategoryID: rGetArmorCategoryBySlugRequest.GetArmorCategory().GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Rimuovo equipaggiamento attuale
			if rGetPlayerArmorEquippedByCategoryID.GetArmor().GetID() > 0 {
				if _, err = config.App.Server.Connection.EquipArmor(helpers.NewContext(1), &pb.EquipArmorRequest{
					PlayerID: c.Player.ID,
					ArmorID:  rGetPlayerArmorEquippedByCategoryID.GetArmor().GetID(),
					Equip:    false,
				}); err != nil {
					c.Logger.Panic(err)
				}
			}

			// Aggiorno con quello nuovo
			if _, err = config.App.Server.Connection.EquipArmor(helpers.NewContext(1), &pb.EquipArmorRequest{
				PlayerID: c.Player.ID,
				ArmorID:  c.Payload.EquipID,
				Equip:    false,
			}); err != nil {
				c.Logger.Panic(err)
			}
		case "weapons":
			// Recupero arma attualmente equipaggiata
			var rGetPlayerWeaponEquipped *pb.GetPlayerWeaponEquippedResponse
			if rGetPlayerWeaponEquipped, err = config.App.Server.Connection.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Rimovo arma attualmente equipaggiata
			if rGetPlayerWeaponEquipped.GetWeapon().GetID() > 0 {
				if _, err = config.App.Server.Connection.EquipWeapon(helpers.NewContext(1), &pb.EquipWeaponRequest{
					PlayerID: c.Player.ID,
					WeaponID: rGetPlayerWeaponEquipped.GetWeapon().GetID(),
					Equip:    false,
				}); err != nil {
					c.Logger.Panic(err)
				}
			}

			// Aggiorno equipped
			if _, err = config.App.Server.Connection.EquipWeapon(helpers.NewContext(1), &pb.EquipWeaponRequest{
				PlayerID: c.Player.ID,
				WeaponID: c.Payload.EquipID,
				Equip:    false,
			}); err != nil {
				c.Logger.Panic(err)
			}
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "inventory.equip.completed"),
		)
		msg.ParseMode = "markdown"

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true

		// ###################
		// TUTORIAL - Solo il player si trova dentro il tutorial forzo di tornarare al menu
		// ###################
		if c.InTutorial() {
			c.Configurations.ControllerBack.To = &MenuController{}
		}
	}

	return
}
