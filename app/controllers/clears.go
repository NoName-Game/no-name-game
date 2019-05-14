package controllers

import (
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// Clears - Delete redist state and remove row from DB.
func Clears(update tgbotapi.Update) {
	helpers.DeleteRedisAndDbState(helpers.Player)

	message := update.Message
	msg := services.NewMessage(message.Chat.ID, "LOG: Deleted redis/DB row.")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	services.SendMessage(msg)
}
