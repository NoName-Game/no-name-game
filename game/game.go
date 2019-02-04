package game

import (
	"encoding/json"
	"log"
	"strconv"

	"bitbucket.org/no-name-game/no-name/bot"
	"bitbucket.org/no-name-game/no-name/config"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	player Player
)

func init() {
	bootstrap()
}

// Run - The Game!
func Run() {
	updates := bot.GetUpdates()
	for update := range updates {

		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if update.Message.From.UserName == "" {
			msg := bot.NewMessage(update.Message.Chat.ID, "Non puoi continuare senza useraname")
			bot.SendMessage(msg)
			continue
		}

		// Block to test DB
		// First or create
		player.findByUsername(update.Message.From.UserName)
		if player.ID < 1 {
			player = Player{Username: update.Message.From.UserName}
			player.create()
		}

		log.Println(player)
		// ***************************

		testMultistate(update.Message)

		msg := bot.NewMessage(update.Message.Chat.ID, update.Message.Text)
		bot.SendMessage(msg)
	}
}

// Multistate
func testMultistate(message *tgbotapi.Message) {
	var state PlayerState
	config.Database.Where("player_id = ?", message.From.ID).First(&state)

	if state.ID < 1 {
		state = PlayerState{PlayerID: message.From.ID, Function: "Tutorial"}
		state.create()
	}

	// DA RICONTROLLARE
	if state.Function != "" && state.Function != "Start" {
		switch state.Function {
		case "Sintesi":
			sintesi(message, state)
		}
	} else {
		switch message.Text {
		case "Sintesi":
			sintesi(message, state)
		}
	}
	///////////////////////////
}

func sintesi(message *tgbotapi.Message, playerState PlayerState) {
	//Payload function
	type functionPayload struct {
		Rosso int
		Blu   int
	}

	var payloadPLayer functionPayload
	rawPayload := []byte(playerState.Payload)
	err := json.Unmarshal(rawPayload, &payloadPLayer)
	if err != nil {
		// error back to menu
	}

	switch playerState.Stage {
	case 0:
		playerState.Function = "Sintesi"
		playerState.Stage = 1
		payloadUpdated, _ := json.Marshal(functionPayload{})
		playerState.Payload = string(payloadUpdated)

		playerState.update()

		msg := bot.NewMessage(message.Chat.ID, "Ho solo settato lo state ora, quanto mana BLU vuoi?")
		bot.SendMessage(msg)

	case 1:
		//Mana Blu
		payloadPLayer.Blu, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payloadPLayer)
		playerState.Payload = string(payloadUpdated)
		playerState.Stage = 2

		playerState.update()

		msg := bot.NewMessage(message.Chat.ID, "quanto mana ROSSO vuoi?")
		bot.SendMessage(msg)
	case 2:
		//Mana Rosso
		payloadPLayer.Rosso, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payloadPLayer)
		playerState.Payload = string(payloadUpdated)
		playerState.Stage = 3

		playerState.update()

		msg := bot.NewMessage(message.Chat.ID, "Sei sicuro di voler concludere?")
		bot.SendMessage(msg)
	case 3:
		playerState.Function = "Start"
		playerState.Payload = string("")
		playerState.Stage = 0

		playerState.update()

		msg := bot.NewMessage(message.Chat.ID, "Bravo hai concluso ora puoi andare al'inizio.")
		bot.SendMessage(msg)
	}
}
