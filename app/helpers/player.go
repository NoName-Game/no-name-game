package helpers

import (
	"errors"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	//===================================
	// Public
	Player nnsdk.Player
	//=====================================
)

// HandleUser - Check if user exist in DB, if not exist create!
func HandleUser(update tgbotapi.Update) bool {
	// ************************
	// Switch by message type
	// ************************
	var user *tgbotapi.User
	if update.Message != nil {
		user = update.Message.From
	} else if update.CallbackQuery != nil {
		user = update.CallbackQuery.From
	} else {
		return false
	}

	// ************************
	// Check if have username
	// ************************
	if user.UserName == "" {
		msg := services.NewMessage(update.Message.Chat.ID, Trans("miss_username"))
		services.SendMessage(msg)
		return false
	}

	// ************************
	// Check if player exists
	// ************************
	Player, _ = providers.FindPlayerByUsername(user.UserName)

	// If Player does not exists, create!
	if Player.ID < 1 {
		language, _ := providers.FindLanguageBySlug("en")

		Player, _ = providers.CreatePlayer(nnsdk.Player{
			Username: user.UserName,
			ChatID:   int64(user.ID),
			Language: language,
			Inventory: nnsdk.Inventory{
				Items: "{}",
			},
			Stats: nnsdk.PlayerStats{
				AbilityPoint: 1,
			},
		})

		return true
	}

	// ************************
	// Check if player is die
	// ************************
	if _, err := GetPlayerStateByFunction(Player, "route.death"); err == nil {
		// controllers.PlayerDeath(update) TODO: FIXME
		return false
	}

	return true
}

// GetPlayerStateByFunction - Check if function exist in player states
func GetPlayerStateByFunction(player nnsdk.Player, function string) (playerState nnsdk.PlayerState, err error) {
	for i, state := range player.States {
		if state.Function == function {
			playerState = player.States[i]
			return playerState, nil
		}
	}

	return playerState, errors.New("State not found!")
}

// CheckPlayerHaveOneEquippedWeapon
// Verifica se il player ha almeno un'arma equipaggiata
func CheckPlayerHaveOneEquippedWeapon(player nnsdk.Player) bool {
	for _, weapon := range player.Weapons {
		if *weapon.Equipped == true {
			return true
		}
	}

	return false
}
