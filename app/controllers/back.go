package controllers

import (
	"os"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Back delete only redis state, but not delete state stored in DB.
func Back(update tgbotapi.Update) {
	helpers.DelRedisState(helpers.Player)

	message := update.Message
	if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
		msg := services.NewMessage(message.Chat.ID, "LOG: Deleted only redis state without completion.")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		services.SendMessage(msg)
	}
	Menu(update)
}
