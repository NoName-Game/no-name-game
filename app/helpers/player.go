package helpers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// HandleUser - Check if user exist in DB, if not exist create!
func HandleUser(user *tgbotapi.User) bool {
	Player, _ = providers.FindPlayerByUsername(user.UserName)

	if Player.ID < 1 {
		language, _ := providers.FindLanguageBySlug("en")

		newPlayer := nnsdk.Player{
			Username: user.UserName,
			ChatID:   int64(user.ID),
			Language: language,
			Inventory: nnsdk.Inventory{
				Items: "{}",
			},
			Stats: nnsdk.PlayerStats{
				AbilityPoint: 1,
			},
		}

		// 1 - Create new player
		Player, _ = providers.CreatePlayer(newPlayer)
	}

	return true
}

// GetPlayerStateByFunction - Check if function exist in player states
func GetPlayerStateByFunction(player nnsdk.Player, function string) (playerState nnsdk.PlayerState) {
	for i, state := range player.States {
		if state.Function == function {
			playerState = player.States[i]
		}
	}

	return
}
