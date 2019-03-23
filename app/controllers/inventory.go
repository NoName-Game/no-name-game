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
		//
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
			helpers.Trans("inventory.summary", player.Language.Slug),
			helpers.Trans("inventory.equip", player.Language.Slug),
			helpers.Trans("inventory.destroy", player.Language.Slug),
		}) {
			state.Stage = 1
			state.Update()
			validationFlag = true
		}
	case 1:
		validationFlag = true
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
	// 1 - Inventory summary
	// 2 -
	//====================================
	switch state.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(inventoryPayload{})
		state.Payload = string(payloadUpdated)
		state.Update()

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.intro", player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("inventory.summary", player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("inventory.equip", player.Language.Slug)),
				tgbotapi.NewKeyboardButton(helpers.Trans("inventory.destroy", player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
				tgbotapi.NewKeyboardButton("clears"),
			),
		)
		services.SendMessage(msg)
	case 1:
		// If is valid input
		// if validationFlag {
		// 	payload.Manager = message.Text
		// 	payloadUpdated, _ := json.Marshal(payload)
		// 	state.Payload = string(payloadUpdated)
		// 	state.Update()
		// }

		switch message.Text {
		case helpers.Trans("inventory.summary", player.Language.Slug):
			var recap string

			// Summary Resources
			recap += "\n" + helpers.Trans("resources", player.Language.Slug) + ":\n"
			playerResources := player.Inventory.ToMap()
			for r, q := range playerResources {
				recap += "- " + models.GetResourceByID(r).Name + " (" + (strconv.Itoa(q)) + ")\n"
			}

			// Summary Weapons
			recap += "\n" + helpers.Trans("weapons", player.Language.Slug) + ":\n"
			for _, weapon := range player.Weapons {
				recap += "- " + weapon.Name + "\n"
			}

			// Summary Armors
			recap += "\n" + helpers.Trans("armors", player.Language.Slug) + ":\n"
			for _, armor := range player.Armors {
				recap += "- " + armor.Name + "\n"
			}

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.recap", player.Language.Slug)+recap)
			services.SendMessage(msg)
		case helpers.Trans("inventory.equip", player.Language.Slug):
			inventoryEquip(update, player)
		}
	}
}

func inventoryEquip(update tgbotapi.Update, player models.Player) {
	message := update.Message

	msg := services.NewMessage(message.Chat.ID, "Bella")
	services.SendMessage(msg)
}
