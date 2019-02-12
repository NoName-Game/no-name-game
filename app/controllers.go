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

func backAll(update tgbotapi.Update) {
	delRedisState(player)

	message := update.Message
	msg := services.NewMessage(message.Chat.ID, "Home")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	services.SendMessage(msg)
}

// Only for testing multi-stage
func testMultiStage(update tgbotapi.Update) {
	message := update.Message
	routeName := "test-multi-stage"

	state := player.getStateByFunction(routeName)
	if state.ID < 1 {
		state.Function = routeName
		state.PlayerID = player.ID
		state.create()
	}
	setRedisState(player, routeName)

	switch state.Stage {
	case 0:
		state.Stage = 1
		state.update()

		msg := services.NewMessage(message.Chat.ID, "This is stage 0.")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Go to stage 1"),
			),
		)
		services.SendMessage(msg)

	case 1:
		state.delete()
		delRedisState(player)

		msg := services.NewMessage(message.Chat.ID, "This is stage 1. Bye.")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
			),
		)
		services.SendMessage(msg)
	}

}

// Only for testing multi-state
func testMultiState(update tgbotapi.Update) {
	message := update.Message
	routeName := "test-multi-state"
	type payloadStruct struct {
		Red   int
		Green int
		Blue  int
	}

	state := player.getStateByFunction(routeName)
	if state.ID < 1 {
		state.Function = routeName
		state.PlayerID = player.ID
		state.create()
	}
	setRedisState(player, routeName)

	//FIXME: Partendo da qui, andare a cercare tramite metod se il player ha uno stato per questa funzione
	// Aggiungere una colonna che indichi che questo status è bloccante e non è pobbisibile avviare altri stati
	// Aggiungere una colonna che indichi tra quanto tempo è possibile accedere a questa funzione

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
		delRedisState(player)

		msg := services.NewMessage(message.Chat.ID, "End")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
			),
		)
		services.SendMessage(msg)
	}
}

// EsterEgg for debug
func theAnswerIs(update tgbotapi.Update) {
	log.Println(42)
}
