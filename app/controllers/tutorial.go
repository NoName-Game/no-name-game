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
			SendMessages(helpers.GenerateTextArray(routeName), 2*time.Second)
		}
	}
	//====================================
}

// sendMessages - Send multiple message every elapsedTime.
func SendMessages(texts []string, elapsedTime time.Duration) {
	lastMessage := services.SendMessage(services.NewMessage(helpers.Player.ChatID, texts[0]))
	var previousText string
	for i := 1; i < 3; i++ {
		time.Sleep(elapsedTime)
		/*previousText = */ services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, lastMessage.MessageID, texts[i])) //.Text
	}
	for i := 3; i < len(texts); i++ {
		time.Sleep(elapsedTime)
		previousText = services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, lastMessage.MessageID, previousText+"\n"+texts[i])).Text
	}
}
