package controllers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"time"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/provider"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// TimedMessages - Send multiple messages with a delay.
func TimedMessages(texts []string, toChatID int64, seconds time.Duration) {
	for _, text := range texts {
		services.SendMessage(services.NewMessage(toChatID, text))
		time.Sleep(seconds)
	}
}

// StartTutorial - This is the first command called from telegram when bot started.
func StartTutorial(update tgbotapi.Update) {
	message := update.Message
	routeName := "route.start"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := "Wrong input, please repeat or exit."
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
			//========================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, helpers.Player)
			//====================================

			textToSend := helpers.Trans("complete")
			msg := services.NewMessage(message.Chat.ID, textToSend)
			text, _ := services.GetTranslation("start_game", player.Language.Slug)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(text)))
			services.SendMessage(msg)
		}
	case 2:
		//========================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(state, player)
		//====================================
		TimedMessages(services.GenerateTextArray("tutorial", player.Language.Slug), message.Chat.ID, 1*time.Second)
	}
	//====================================
}
