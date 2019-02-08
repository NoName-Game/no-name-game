package game

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/bot"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func getPlayerStatePayload(player *Player, funcInterface interface{}) {
	err := json.Unmarshal([]byte(player.State.Payload), &funcInterface)
	if err != nil {
		// error back to menu
	}
}

func setPlayerState(player *Player, state PlayerState) {
	player.State.Function = state.Function
	player.State.Stage = state.Stage
	player.State.Payload = state.Payload

	player.State.update()
}

func testMultiState(update tgbotapi.Update) {
	message := update.Message

	//Payload function
	type functionPayload struct {
		Rosso int
		Blu   int
	}

	var payloadPLayer functionPayload
	getPlayerStatePayload(&player, &payloadPLayer)

	switch player.State.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(functionPayload{})
		setPlayerState(&player, PlayerState{
			Function: "test-multi-state",
			Stage:    1,
			Payload:  string(payloadUpdated),
		})

		msg := bot.NewMessage(message.Chat.ID, "Ho solo settato lo state ora, quanto mana BLU vuoi?")
		bot.SendMessage(msg)

	case 1:
		//Mana Blu
		payloadPLayer.Blu, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payloadPLayer)

		setPlayerState(&player, PlayerState{
			Function: "test-multi-state",
			Stage:    2,
			Payload:  string(payloadUpdated),
		})

		msg := bot.NewMessage(message.Chat.ID, "quanto mana ROSSO vuoi?")
		bot.SendMessage(msg)
	case 2:
		//Mana Rosso
		payloadPLayer.Rosso, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payloadPLayer)

		setPlayerState(&player, PlayerState{
			Function: "test-multi-state",
			Stage:    3,
			Payload:  string(payloadUpdated),
		})

		msg := bot.NewMessage(message.Chat.ID, "Sei sicuro di voler concludere?")
		bot.SendMessage(msg)
	case 3:
		setPlayerState(&player, PlayerState{})

		msg := bot.NewMessage(message.Chat.ID, "Bravo hai concluso ora puoi andare al'inizio.")
		bot.SendMessage(msg)
	}
}
