package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// CheckUser - Check if user exist in DB, if not exist create!
func CheckUser(message *tgbotapi.Message) (player models.Player) {
	player = models.FindPlayerByUsername(message.From.UserName)
	if player.ID < 1 {
		// New Player Inventory
		newInventory := models.Inventory{Items: ""}
		newInventory.Create()

		player = models.Player{
			Username:  message.From.UserName,
			ChatID:    message.Chat.ID,
			Language:  models.GetLangBySlug("en"),
			Inventory: newInventory,
		}

		// 1 - Create new player
		player.Create()

		// 2- New galaxy chunk and return player star
		playerStar := NewGalaxyChunk(int(player.ID))

		// 3 - Add first star to player
		player.AddStar(playerStar)

		// 4 - Add and create ship
		ship := NewStartShip()
		player.AddShip(ship)

		// 5 - Register first player position
		player.AddPosition(models.PlayerPosition{
			X: playerStar.X,
			Y: playerStar.Y,
			Z: playerStar.Z,
		})
	}

	return
}
