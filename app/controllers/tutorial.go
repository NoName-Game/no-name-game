package controllers

import (
	"time"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/provider"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StartTutorial - This is the first command called from telegram when bot started.
func StartTutorial(update tgbotapi.Update) {
	message := update.Message
	routeName := "route.start"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage")
	switch state.Stage {
	case 1:
		lang, err := provider.FindLanguageBy(message.Text, "name")
		if err != nil {
			services.ErrorHandler("Cant find language", err)
		}

		if lang.ID >= 1 {
			validationFlag = true
		}

		{
			_, err := provider.UpdatePlayer(nnsdk.Player{ID: helpers.Player.ID, LanguageID: lang.ID})
			if err != nil {
				services.ErrorHandler("Cant update player", err)
			}
		}
	case 2:
		if message.Text == helpers.Trans("route.start.openEye") {
			validationFlag = true
		}
	case 3:
		validationMessage = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if helpers.GetPlayerStateByFunction(helpers.Player, "route.mission") == (nnsdk.PlayerState{}) {
			validationFlag = true
		}
	case 4:
		validationMessage = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if helpers.GetPlayerStateByFunction(helpers.Player, "route.crafting") == (nnsdk.PlayerState{}) {
			validationFlag = true
		}
	case 5:
		validationMessage = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if helpers.GetPlayerStateByFunction(helpers.Player, "route.inventory.equip") == (nnsdk.PlayerState{}) {
			validationFlag = true
		}
	case 6:
		validationMessage = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if helpers.GetPlayerStateByFunction(helpers.Player, "route.hunting") == (nnsdk.PlayerState{}) {
			validationFlag = true
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			services.SendMessage(validatorMsg)
		}
	}

	//====================================
	// Stage
	//====================================
	//====================================
	// Language -> Messages -> Exploration -> Crafting -> Hunting
	//====================================
	switch state.Stage {
	case 0:
		msg := services.NewMessage(message.Chat.ID, "Select language")

		languages, err := provider.GetLanguages()
		if err != nil {
			services.ErrorHandler("Cant get languages", err)
		}

		keyboard := make([]tgbotapi.KeyboardButton, len(languages))
		for i, lang := range languages {
			keyboard[i] = tgbotapi.NewKeyboardButton(lang.Name)
		}

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
		state.Stage = 1
		state, _ = provider.UpdatePlayerState(state)
		services.SendMessage(msg)
	case 1:
		if validationFlag {
			// Messages
			texts := helpers.GenerateTextArray(routeName)
			msg := services.NewMessage(helpers.Player.ChatID, texts[0])
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			lastMessage := services.SendMessage(msg)
			var previousText string
			for i := 1; i < 3; i++ {
				time.Sleep(2 * time.Second)
				services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, lastMessage.MessageID, texts[i])) //.Text
			}
			for i := 3; i < 12; i++ {
				time.Sleep(2 * time.Second)
				previousText = services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, lastMessage.MessageID, previousText+"\n"+texts[i])).Text
			}
			lastMessage = services.SendMessage(services.NewMessage(helpers.Player.ChatID, texts[12]))
			previousText = lastMessage.Text
			for i := 13; i < len(texts); i++ {
				time.Sleep(time.Second)
				previousText = services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, lastMessage.MessageID, previousText+"\n"+texts[i])).Text
			}
			edit := services.NewEditMessage(helpers.Player.ChatID, lastMessage.MessageID, helpers.Trans("route.start.explosion"))
			edit.ParseMode = "HTML"
			services.SendMessage(edit)
			msg = services.NewMessage(helpers.Player.ChatID, "...")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("route.start.openEye"))))
			services.SendMessage(msg)
			state.Stage = 2
			state, _ = provider.UpdatePlayerState(state)
		}
	case 2:
		if validationFlag {
			// First Exploration
			services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstExploration")))
			state.Stage = 3
			state, _ = provider.UpdatePlayerState(state)
			StartMission(update)
		}
	case 3:
		if validationFlag {
			// First Crafting
			services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstCrafting")))
			state.Stage = 4
			state, _ = provider.UpdatePlayerState(state)
			Crafting(update)
		}
	case 4:
		if validationFlag {
			// Equip weapon
			services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstWeaponEquipped")))
			state.Stage = 5
			state, _ = provider.UpdatePlayerState(state)
			InventoryEquip(update)
		}
	case 5:
		if validationFlag {
			services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstHunting")))
			state.Stage = 6
			state, _ = provider.UpdatePlayerState(state)
			Hunting(update)
		}
	case 6:
		if validationFlag {
			helpers.FinishAndCompleteState(state, helpers.Player)
		}
	}
	//====================================
}
