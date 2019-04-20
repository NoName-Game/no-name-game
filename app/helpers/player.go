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
		player, _ = provider.CreatePlayer(newPlayer)
	}

	return
}

func GetPlayerStateByFunction(player nnsdk.Player, function string) (playerState nnsdk.PlayerState) {
	playerStates, err := provider.GetPlayerStates(player)
	if err != nil {
		panic(err)
	}

	for i, state := range playerStates {
		if state.Function == function {
			playerState = playerStates[i]
		}
	}

	return
}
