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
			Username:  message.From.UserName,
			Language:  models.GetLangBySlug("en"),
			Inventory: models.Inventory{Items: ""},
		}

		// 1 - Create new player
		player.Create()

		// 2- New galaxy chunk and return player star
		var playerStar models.Star
		playerStar = NewGalaxyChunk(int(player.ID))

		// 3 - Add first star to player
		player.AddStar(playerStar)

		// 4 - Register first player position
		player.AddPosition(models.PlayerPosition{
			X: playerStar.X,
			Y: playerStar.Y,
			Z: playerStar.Z,
		})
	}

	return
}
