package controllers

import (
	"log"
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
			helpers.FinishAndCompleteState(state, helpers.Player)
			// Messages
			log.Println("Ehy, ....")
			SendMultipleMessages(helpers.GenerateTextArray(routeName), 1*time.Second)
		}
	}
	//====================================
}

// sendMultipleMessages - Send multiple message every elapsedTime.
func SendMultipleMessages(texts []string, elapsedTime time.Duration) {
	log.Println(texts)
	for _, text := range texts {
		time.Sleep(elapsedTime)
		services.SendMessage(services.NewMessage(helpers.Player.ChatID, text))
	}
}
