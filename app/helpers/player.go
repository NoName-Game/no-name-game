package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/app/provider"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// CheckUser - Check if user exist in DB, if not exist create!
func CheckUser(message *tgbotapi.Message) (player nnsdk.Player) {
	player, _ = provider.FindPlayerByUsername(message.From.UserName)
	language, _ := provider.FindLanguageBySlug("en")

	if player.ID < 1 {
		newPlayer := nnsdk.Player{
			Username: message.From.UserName,
			ChatID:   message.Chat.ID,
			Language: language,
			Inventory: nnsdk.Inventory{
				Items: "",
			},
			Stats: nnsdk.PlayerStats{},
		}

		// 1 - Create new player
		player, _ := provider.CreatePlayer(newPlayer)

		panic(player)

		//TODO: Continue here, move al login in WS like new registrer command

		// 2- New galaxy chunk and return player star
		// playerStar := NewGalaxyChunk(int(player.ID))

		// // 3 - Add first star to player
		// player.AddStar(playerStar)

		// // 4 - Add and create ship
		// ship := NewStartShip()
		// player.AddShip(ship)

		// // 5 - Register first player position
		// player.AddPosition(models.PlayerPosition{
		// 	X: playerStar.X,
		// 	Y: playerStar.Y,
		// 	Z: playerStar.Z,
		// })
	}

	return
}
