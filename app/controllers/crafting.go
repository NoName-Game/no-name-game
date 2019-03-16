package controllers

import (
	"encoding/json"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// Crafting
func Crafting(update tgbotapi.Update, player models.Player) {
	//====================================
	// Init Func!
	//====================================
	type payloadCrafting struct {
		Item     string
		Category string
	}
	message := update.Message
	routeName := "crafting"
	state := helpers.StartAndCreatePlayerState(routeName, player)
	var payload payloadCrafting
	helpers.UnmarshalPayload(state.Payload, &payload)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := "Wrong input, please repeat or exit."
	switch state.Stage {
	case 0:
		if helpers.InArray(message.Text, []string{"Armors", "Weapons"}) {
			state.Stage = 1
			state.Update()
			validationFlag = true
		}
	case 1:
		if helpers.InArray(message.Text, helpers.GetAllCategories()) {
			state.Stage = 2
			state.Update()
			validationFlag = true
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			services.SendMessage(validatorMsg)
		}
	}

	// Logic flux
	//		0		1		 	2
	// -> What -> Category -> Resources

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(payloadCrafting{})
		state.Payload = string(payloadUpdated)
		state.Update()

		msg := services.NewMessage(message.Chat.ID, "What do you want craft?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Armors"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Weapons"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
				tgbotapi.NewKeyboardButton("clears"),
			),
		)
		services.SendMessage(msg)
	case 1:
		// If is valid input
		if validationFlag {
			payload.Item = helpers.Slugger(message.Text)
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Update()
		}

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch payload.Item {
		case "armors":
			armorCategories := models.GetAllArmorCategories()
			for _, category := range armorCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(category.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case "weapons":
			weaponCategories := models.GetAllWeaponCategories()
			for _, category := range weaponCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(category.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("back"),
			tgbotapi.NewKeyboardButton("clears"),
		))

		msg := services.NewMessage(message.Chat.ID, "Which type?")
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)

	case 2:
		// If is valid input
		if validationFlag {
			payload.Category = helpers.Slugger(message.Text)
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Update()
		}

		//ONLY FOR DEBUG - Add one resource
		debugItem := models.GetResourceByID(42)
		player.Inventory.AddResource(debugItem, 2)

		var keyboardRowResources [][]tgbotapi.KeyboardButton
		playerResources := player.Inventory.ToMap()
		for _, resource := range playerResources {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(resource))
			keyboardRowResources = append(keyboardRowResources, keyboardRow)
		}

		// Clear and exit
		keyboardRowResources = append(keyboardRowResources, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("back"),
			tgbotapi.NewKeyboardButton("clears"),
		))

		msg := services.NewMessage(message.Chat.ID, "Choose which resources to use")
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowResources,
		}
		services.SendMessage(msg)

		//FIXME: continue me :)
	case 80:
		// If is valid input
		if validationFlag {
			payload.Item = message.Text
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Update()
		}

		msg := services.NewMessage(message.Chat.ID, "Finish?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("YES!"),
				tgbotapi.NewKeyboardButton("Wrong answare!"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
			),
		)
		services.SendMessage(msg)
	case 90:
		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(state, player)
		//====================================

		msg := services.NewMessage(message.Chat.ID, "Completed! :)")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
			),
		)
		services.SendMessage(msg)
	}
}
