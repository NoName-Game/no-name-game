package controllers

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Inventory
func Inventory(update tgbotapi.Update, player models.Player) {
	//====================================
	// Init Func!
	//====================================
	type inventoryPayload struct {
		Manager string
	}

	message := update.Message
	routeName := "inventory"
	state := helpers.StartAndCreatePlayerState(routeName, player)
	var payload inventoryPayload
	helpers.UnmarshalPayload(state.Payload, &payload)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage", player.Language.Slug)
	switch state.Stage {
	case 0:
		if helpers.InArray(message.Text, []string{
			helpers.Trans("resources", player.Language.Slug),
			helpers.Trans("armors", player.Language.Slug),
			helpers.Trans("weapons", player.Language.Slug),
		}) {
			state.Stage = 1
			state.Update()
			validationFlag = true
		}
	case 1:
		// if helpers.InArray(message.Text, helpers.GetAllTranslatedSlugCategoriesByLocale(player.Language.Slug)) {
		// 	state.Stage = 2
		// 	state.Update()
		// 	validationFlag = true
		// }
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			validatorMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			services.SendMessage(validatorMsg)
		}
	}

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(inventoryPayload{})
		state.Payload = string(payloadUpdated)
		state.Update()

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.intro", player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("resources", player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("armors", player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("weapons", player.Language.Slug)),
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
			payload.Manager = message.Text
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Update()
		}

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch payload.Manager {
		case helpers.Trans("resources", player.Language.Slug):
			playerResources := player.Inventory.ToMap()
			for r, q := range playerResources {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(models.GetResourceByID(r).Name + " (" + (strconv.Itoa(q)) + ")"))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("armors", player.Language.Slug):
			for _, armor := range player.Armors {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(armor.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons", player.Language.Slug):
			for _, weapon := range player.Weapons {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(weapon.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("back"),
			tgbotapi.NewKeyboardButton("clears"),
		))

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.recap", player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)
	}
}
