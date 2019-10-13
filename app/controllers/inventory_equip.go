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
// Inventory
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
	c.RouteName = "route.inventory.equip"
	c.Update = update
	c.Message = update.Message

	// Check current state for this routes
	state, isNewState := helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// Set and load payload
	helpers.UnmarshalPayload(state.Payload, c.Payload)

	// It's first message
	if isNewState {
		c.Stage(state)
		return
	}

	// Go to validator
	c.Validation.HasErrors, state = c.Validator(state)
	if !c.Validation.HasErrors {
		state, _ = providers.UpdatePlayerState(state)
		c.Stage(state)
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
func (c *InventoryEquipController) Validator(state nnsdk.PlayerState) (hasErrors bool, newState nnsdk.PlayerState) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch state.Stage {
	case 0:
		if helpers.InArray(c.Message.Text, []string{
			helpers.Trans("armors"),
			helpers.Trans("weapons"),
		}) {
			state.Stage = 1
			return false, state
		}
	case 1:
		if strings.Contains(c.Message.Text, helpers.Trans("equip")) {
			state.Stage = 2
			return false, state
		}
	case 2:
		if c.Message.Text == helpers.Trans("confirm") {
			state.Stage = 3
			return false, state
		}
	}

	return true, state
}

//====================================
// Stage
//====================================
func (c *InventoryEquipController) Stage(state nnsdk.PlayerState) {
	switch state.Stage {
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
		// If is valid input
		c.Payload.Type = c.Message.Text
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, _ = providers.UpdatePlayerState(c.State)

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch c.Payload.Type {
		case helpers.Trans("armors"):
			armors, err := providers.GetPlayerArmors(helpers.Player, "false")
			if err != nil {
				services.ErrorHandler("Cant get player armors", err)
			}

			// Each player armors
			for _, armor := range armors {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("equip") + " " + armor.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons"):
			weapons, err := providers.GetPlayerWeapons(helpers.Player, "false")
			if err != nil {
				services.ErrorHandler("Cant get player weapons", err)
			}

			// Each player weapons
			for _, weapon := range weapons {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("equip") + " " + weapon.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		))

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.what"))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)
	case 2:
		var equipmentName string

		// Clear text from Add and other shit.
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

		c.Payload.EquipID = equipmentID
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, _ = providers.UpdatePlayerState(c.State)

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
	case 3:
		switch c.Payload.Type {
		case helpers.Trans("armors"):
			equipment, err := providers.GetArmorByID(c.Payload.EquipID)
			if err != nil {
				services.ErrorHandler("Cant find armor by ID", err)
			}

			// Stupid poninter stupid json pff
			t := new(bool)
			*t = true
			equipment.Equipped = t

			_, err = providers.UpdateArmor(equipment)
			if err != nil {
				services.ErrorHandler("Cant update armor", err)
			}
		case helpers.Trans("weapons"):
			equipment, err := providers.GetWeaponByID(c.Payload.EquipID)
			if err != nil {
				services.ErrorHandler("Cant find weapon by ID", err)
			}

			// Stupid poninter stupid json pff
			t := new(bool)
			*t = true
			equipment.Equipped = t

			_, err = providers.UpdateWeapon(equipment)
			if err != nil {
				services.ErrorHandler("Cant update weapon", err)
			}
		}

		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(state, helpers.Player)
		//====================================

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.equip.completed"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)
	}
}
