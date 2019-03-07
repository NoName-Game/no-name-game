package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// CheckUser - Check if user exist in DB, if not exist create!
func CheckUser(message *tgbotapi.Message) (player models.Player) {
	player = models.FindPlayerByUsername(message.From.UserName)
	if player.ID < 1 {
		player = models.Player{
			Username:  message.From.UserName,
			ChatID:    message.Chat.ID,
			Language:  models.GetLangBySlug("en"),
			Inventory: models.Inventory{Items: ""},
		}

		// Create new player
		player.Create()

		// New galaxy chunk
		NewGalaxyChunk(int(player.ID))
	}

	return
}
