package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/app/provider"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// HandleUser - Check if user exist in DB, if not exist create!
func HandleUser(user *tgbotapi.User) bool {
	Player, _ = provider.FindPlayerByUsername(user.UserName)

	if Player.ID < 1 {
		language, _ := provider.FindLanguageBySlug("en")

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
		Player, _ = provider.CreatePlayer(newPlayer)
	}

	return true
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
