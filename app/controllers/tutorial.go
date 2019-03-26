package controllers

import (
	"time"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
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
func StartTutorial(update tgbotapi.Update, player models.Player) {
	message := update.Message
	routeName := "start"
	state := helpers.StartAndCreatePlayerState(routeName, player)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := "Wrong input, please repeat or exit."
	switch state.Stage {
	case 1:
		if lang := models.GetLangByValue(message.Text); lang.Value != "" {
			//Il languaggio esiste.
			validationFlag = true
			player.Language = lang
			player.Update()
		}
	case 2:
		if text, _ := services.GetTranslation("start_game", player.Language.Slug); text == message.Text {
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
	switch state.Stage {
	case 0:
		msg := services.NewMessage(message.Chat.ID, "Select language")
		keyboard := make([]tgbotapi.KeyboardButton, len(models.GetAllLangs()))
		for i, lang := range models.GetAllLangs() {
			keyboard[i] = tgbotapi.NewKeyboardButton(lang.Value)
		}
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
		state.Stage = 1
		state.Update()
		services.SendMessage(msg)
	case 1:
		if true == validationFlag {
			state.Stage = 2
			state.Update()
			textToSend, _ := services.GetTranslation("complete", player.Language.Slug)
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
