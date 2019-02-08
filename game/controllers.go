package game

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/bot"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func setPlayerState() {

}

func testMultiState(update tgbotapi.Update) {
	message := update.Message

	//Payload function
	type functionPayload struct {
		Rosso int
		Blu   int
	}

	var payloadPLayer functionPayload
	rawPayload := []byte(player.State.Payload)
	err := json.Unmarshal(rawPayload, &payloadPLayer)
	if err != nil {
		// error back to menu
	}

	switch player.State.Stage {
	case 0:
		player.State.Function = "Sintesi"
		player.State.Stage = 1
		payloadUpdated, _ := json.Marshal(functionPayload{})
		player.State.Payload = string(payloadUpdated)

		player.update()

		msg := bot.NewMessage(message.Chat.ID, "Ho solo settato lo state ora, quanto mana BLU vuoi?")
		bot.SendMessage(msg)

	case 1:
		//Mana Blu
		payloadPLayer.Blu, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payloadPLayer)
		player.State.Payload = string(payloadUpdated)
		player.State.Stage = 2

		player.update()

		msg := bot.NewMessage(message.Chat.ID, "quanto mana ROSSO vuoi?")
		bot.SendMessage(msg)
	case 2:
		//Mana Rosso
		payloadPLayer.Rosso, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payloadPLayer)
		player.State.Payload = string(payloadUpdated)
		player.State.Stage = 3

		player.State.update()

		msg := bot.NewMessage(message.Chat.ID, "Sei sicuro di voler concludere?")
		bot.SendMessage(msg)
	case 3:
		player.State.Function = "Start"
		player.State.Payload = string("")
		player.State.Stage = 0

		player.State.update()

		msg := bot.NewMessage(message.Chat.ID, "Bravo hai concluso ora puoi andare al'inizio.")
		bot.SendMessage(msg)
	}
}
