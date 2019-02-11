package app

import (
	"encoding/json"
	"log"
	"strconv"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

//********************
//
//		TEST
//
//********************

// Only for testing multi-state
func testMultiState(update tgbotapi.Update) {
	message := update.Message

	state := player.getStateByFunction("test-multi-state")
	if state.ID < 1 {
		state.Function = "test-multi-state"
		state.PlayerID = player.ID
		state.create()
	}

	//FIXME:
	err := services.Redis.Set(strconv.FormatUint(uint64(player.ID), 10), "test-multi-state", 0).Err()
	if err != nil {
		panic(err)
	}

	// log.Println("here", state)

	//FIXME: Partendo da qui, andare a cercare tramite metod se il player ha uno stato per questa funzione
	// Aggiungere una colonna che indichi che questo status è bloccante e non è pobbisibile avviare altri stati
	// Aggiungere una colonna che indichi tra quanto tempo è possibile accedere a questa funzione

	//Payload function
	type payloadStruct struct {
		Red   int
		Green int
		Blue  int
	}

	var payload payloadStruct
	unmarshalPayload(state.Payload, &payload)

	switch state.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(payloadStruct{})

		state.Stage = 1
		state.Payload = string(payloadUpdated)
		state.update()

		msg := services.NewMessage(message.Chat.ID, "State setted, How much of R?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
			),
		)
		services.SendMessage(msg)

	case 1:
		//R
		payload.Red, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payload)

		state.Stage = 2
		state.Payload = string(payloadUpdated)
		state.update()

		msg := services.NewMessage(message.Chat.ID, "Stage 2 setted, How much of G?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("0"),
			),
		)
		services.SendMessage(msg)
	case 2:
		//G
		payload.Green, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payload)

		state.Stage = 3
		state.Payload = string(payloadUpdated)
		state.update()

		msg := services.NewMessage(message.Chat.ID, "Stage 2 setted, How much of B?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
			),
		)
		services.SendMessage(msg)
	case 3:
		//B
		payload.Blue, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payload)

		state.Stage = 4
		state.Payload = string(payloadUpdated)
		state.update()

		msg := services.NewMessage(message.Chat.ID, "Finish?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("YES!"),
			),
		)
		services.SendMessage(msg)
	case 4:
		state.delete()

		//FIXME:
		err := services.Redis.Del(strconv.FormatUint(uint64(player.ID), 10)).Err()
		if err != nil {
			panic(err)
		}

		msg := services.NewMessage(message.Chat.ID, "End")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		services.SendMessage(msg)
	}
}

// EsterEgg for debug
func theAnswerIs(update tgbotapi.Update) {
	log.Println(42)
}
