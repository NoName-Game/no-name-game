package controllers

import (
	"time"

	"bitbucket.org/no-name-game/no-name/services"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StartMission - start an exploration
func StartMission(update tgbotapi.Update, player models.Player) {

	//message := update.Message
	routeName := "mission"

	state := helpers.StartAndCreatePlayerState(routeName, player)

	eTypes := [3]string{"sottosuolo", "terreno", "atmosfera"}

	switch state.Stage {
	case 0:
		text, _ := services.GetTranslation("esplorazione", player.Language.Slug)
		msg := services.NewMessage(player.ChatID, text)
		keyboardRow := make([]tgbotapi.KeyboardButton, len(eTypes))
		for i, eType := range eTypes {
			text, _ = services.GetTranslation(eType, player.Language.Slug)
			keyboardRow[i] = tgbotapi.NewKeyboardButton(text)
		}
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboardRow)
		services.SendMessage(msg)
	}

}

func SetFinishTime(hours, minutes, seconds int) (t time.Time) {
	t = time.Now().Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds) + time.Second)
	return
}
