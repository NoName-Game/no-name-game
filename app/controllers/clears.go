package controllers

import (
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// Clears - Delete redist state and remove row from DB.
func Clears(update tgbotapi.Update, player models.Player) {
	// TODO:
	// helpers.DeleteRedisAndDbState(player)

	message := update.Message
	msg := services.NewMessage(message.Chat.ID, "LOG: Deleted redis/DB row.")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	services.SendMessage(msg)
}
