package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/models"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// CheckUser - Check if user exist in DB, if not exist create!
func CheckUser(message *tgbotapi.Message) (player models.Player) {
	player = models.FindPlayerByUsername(message.From.UserName)
	if player.ID < 1 {
		player = models.Player{
			Username: message.From.UserName,
			Language: models.GetDefaultLangID("en"),
		}

		player.Create()
	}

	return
}
