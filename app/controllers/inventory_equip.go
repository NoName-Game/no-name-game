package controllers

import (
	"encoding/json"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//====================================
// InventoryEquipController
//====================================
type InventoryEquipController struct {
	BaseController
	Payload struct {
		Type    string
		EquipID uint
	}
}

//====================================
// Handle
//====================================
func (c *InventoryEquipController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	var err error
	var isNewState bool
	c.RouteName, c.Update, c.Message = "route.inventory.equip", update, update.Message

	// Check current state for this routes
	c.State, isNewState = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// It's first message
	if isNewState {
		c.Stage()
		return
	}

	// Go to validator
	if !c.Validator() {
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}

		// Ok! Run!
		c.Stage()
		return
	}

	// Validator goes errors
	validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
	services.SendMessage(validatorMsg)
	return
}

//====================================
// Validator
//====================================
func (c *InventoryEquipController) Validator() (hasErrors bool) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch c.State.Stage {
	case 0:
		if helpers.InArray(c.Message.Text, []string{
			helpers.Trans("armors"),
			helpers.Trans("weapons"),
		}) {
			c.State.Stage = 1
			return false
		}
	case 1:
		if strings.Contains(c.Message.Text, helpers.Trans("equip")) {
			c.State.Stage = 2
			return false
		}
	case 2:
		if c.Message.Text == helpers.Trans("confirm") {
			c.State.Stage = 3
			return false
		}
	}

	return true
}

//====================================
// Stage
//====================================
func (c *InventoryEquipController) Stage() {
	var err error

	switch c.State.Stage {
	case 0:
		//====================================
		// TODO: RISCRIVIMI!
		//====================================
		currentPlayerEquipment := helpers.Trans("inventory.equip.equipped")

		//////////////////////////////////
		currentPlayerEquipment += "\n" + helpers.Trans("armors") + ":\n"
		eqippedArmors, err := providers.GetPlayerArmors(helpers.Player, "true")
		if err != nil {
			services.ErrorHandler("Cant get equpped player armors", err)
		}

		for _, armor := range eqippedArmors {
			currentPlayerEquipment += "- " + armor.Name
		}
		//////////////////////////////////
		currentPlayerEquipment += "\n\n" + helpers.Trans("weapons") + ":\n"
		eqippedWeapons, err := providers.GetPlayerWeapons(helpers.Player, "true")
		if err != nil {
			services.ErrorHandler("Cant get equpped player weapons", err)
		}

		for _, weapon := range eqippedWeapons {
			currentPlayerEquipment += "- " + weapon.Name
		}
		//////////////////////////////////

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.type")+currentPlayerEquipment)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("armors")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("weapons")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)
	case 1:
		// Costruisco keyboard risposta
		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch c.Payload.Type {
		case helpers.Trans("armors"):
			// Each player armors
			for _, armor := range helpers.Player.Armors {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("equip") + " " + armor.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons"):
			// Ciclo armi player
			for _, weapon := range helpers.Player.Weapons {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("equip") + " " + weapon.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Aggiungo tasti back and clears
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		))

		// Invio messaggio
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.what"))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)

		// Aggiorno stato
		c.Payload.Type = c.Message.Text
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}
	case 2:
		var equipmentName string

		// Ripulisco messaggio
		equipmentName = strings.Split(c.Message.Text, helpers.Trans("equip")+" ")[1]

		var equipmentID uint
		switch c.Payload.Type {
		case helpers.Trans("armors"):
			var armor nnsdk.Armor
			armor, err := providers.FindArmorByName(equipmentName)
			if err != nil {
				services.ErrorHandler("Cant find equip armor name", err)
			}

			equipmentID = armor.ID
		case helpers.Trans("weapons"):
			var weapon nnsdk.Weapon
			weapon, err := providers.FindWeaponByName(equipmentName)
			if err != nil {
				services.ErrorHandler("Cant find equip weapon name", err)
			}

			equipmentID = weapon.ID
		}

		// Invio messaggio per conferma equipaggiamento
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.equip.confirm")+"\n\n "+equipmentName)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)

		// Aggiorno stato
		c.Payload.EquipID = equipmentID
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}
	case 3:
		switch c.Payload.Type {
		case helpers.Trans("armors"):
			equipment, err := providers.GetArmorByID(c.Payload.EquipID)
			if err != nil {
				services.ErrorHandler("Cant find armor by ID", err)
			}

			// Aggiorno equipped
			equipment.Equipped = helpers.SetTrue()
			_, err = providers.UpdateArmor(equipment)
			if err != nil {
				services.ErrorHandler("Cant update armor", err)
			}
		case helpers.Trans("weapons"):
			equipment, err := providers.GetWeaponByID(c.Payload.EquipID)
			if err != nil {
				services.ErrorHandler("Cant find weapon by ID", err)
			}

			// Aggiorno equipped
			equipment.Equipped = helpers.SetTrue()
			_, err = providers.UpdateWeapon(equipment)
			if err != nil {
				services.ErrorHandler("Cant update weapon", err)
			}
		}

		// Invio messaggio
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.equip.completed"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)

		//====================================
		// COMPLETE!
		//====================================
		helpers.FinishAndCompleteState(c.State, helpers.Player)
		//====================================
	}
}
