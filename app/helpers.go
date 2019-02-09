package app

import (
	"encoding/json"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// Helper - Check if user exist in DB, if not exist create!
func checkUser(message *tgbotapi.Message) bool {
	player = findPlayerByUsername(message.From.UserName)
	if player.ID < 1 {
		player = Player{
			Username: message.From.UserName,
		}

		player.create()
	}

	return true
}

// Helper - Unmarshal payload state
func getPlayerStatePayload(player *Player, funcInterface interface{}) {
	err := json.Unmarshal([]byte(player.State.Payload), &funcInterface)
	if err != nil {
		// config.ErrorHandler("Unmarshal payload error"+strconv.FormatUint(uint64(player.State.ID), 10),
		// 	errors.New("Unmarshal payload error"+strconv.FormatUint(uint64(player.State.ID), 10)))
	}
}

// Helper - set state of player in DB
func setPlayerState(player *Player, state PlayerState) {
	player.State.Function = state.Function
	player.State.Stage = state.Stage
	player.State.Payload = state.Payload

	player.update()
}

// trans - late shortCut
func trans(key, locale string, args ...interface{}) (message string) {
	if len(args) <= 0 {
		message, _ = services.GetTranslation(key, locale)
		return
	}

	message, _ = services.GetTranslation(key, locale, args)
	return
}
