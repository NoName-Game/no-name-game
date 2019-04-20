package controllers

import (
	"bitbucket.org/no-name-game/no-name/app/models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// StartTutorial - This is the first command called from telegram when bot started.
func StartTutorial(update tgbotapi.Update, player models.Player) {
	// message := update.Message
	// routeName := "route.start"
	// state := helpers.StartAndCreatePlayerState(routeName, player)

	// //====================================
	// // Validator
	// //====================================
	// validationFlag := false
	// validationMessage := "Wrong input, please repeat or exit."
	// switch state.Stage {
	// case 1:
	// 	if lang := models.GetLangByValue(message.Text); lang.Value != "" {
	// 		//Il languaggio esiste.
	// 		validationFlag = true
	// 		player.Language = lang
	// 		player.Update()
	// 	}
	// }
	// if !validationFlag {
	// 	if state.Stage != 0 {
	// 		validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
	// 		services.SendMessage(validatorMsg)
	// 	}
	// }

	// //====================================
	// // Stage
	// //====================================
	// switch state.Stage {
	// case 0:
	// 	msg := services.NewMessage(message.Chat.ID, "Select language")
	// 	keyboard := make([]tgbotapi.KeyboardButton, len(models.GetAllLangs()))
	// 	for i, lang := range models.GetAllLangs() {
	// 		keyboard[i] = tgbotapi.NewKeyboardButton(lang.Value)
	// 	}
	// 	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
	// 	state.Stage = 1
	// 	state.Update()
	// 	services.SendMessage(msg)
	// case 1:
	// 	if validationFlag {
	// 		//========================
	// 		// IMPORTANT!
	// 		//====================================
	// 		helpers.FinishAndCompleteState(state, player)
	// 		//====================================
	// 		textToSend, _ := services.GetTranslation("complete", player.Language.Slug, nil)
	// 		msg := services.NewMessage(message.Chat.ID, textToSend)
	// 		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	// 		services.SendMessage(msg)
	// 	}
	// }
	//====================================
}
