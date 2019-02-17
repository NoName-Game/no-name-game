package controllers

import (
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// Back delete only redis state, but not delete state stored in DB.
func Back(update tgbotapi.Update, player models.Player) {
	helpers.DelRedisState(player)

	message := update.Message
	msg := services.NewMessage(message.Chat.ID, "LOG: Deleted only redis state without completion.")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	services.SendMessage(msg)
}
