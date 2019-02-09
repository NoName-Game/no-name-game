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

	//Payload function
	type payloadStrct struct {
		Red   int
		Green int
		Blue  int
	}

	var payloadPLayer payloadStrct
	getPlayerStatePayload(&player, &payloadPLayer)

	switch player.State.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(payloadStrct{})

		setPlayerState(&player, PlayerState{
			Function: "test-multi-state",
			Stage:    1,
			Payload:  string(payloadUpdated),
		})

		msg := services.NewMessage(message.Chat.ID, "State setted, How much of R?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
			),
		)
		services.SendMessage(msg)

	case 1:
		//R
		payloadPLayer.Red, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payloadPLayer)

		setPlayerState(&player, PlayerState{
			Function: "test-multi-state",
			Stage:    2,
			Payload:  string(payloadUpdated),
		})

		msg := services.NewMessage(message.Chat.ID, "Stage 2 setted, How much of G?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("0"),
			),
		)
		services.SendMessage(msg)
	case 2:
		//G
		payloadPLayer.Green, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payloadPLayer)

		setPlayerState(&player, PlayerState{
			Function: "test-multi-state",
			Stage:    3,
			Payload:  string(payloadUpdated),
		})

		msg := services.NewMessage(message.Chat.ID, "Stage 2 setted, How much of B?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
			),
		)
		services.SendMessage(msg)
	case 3:
		//B
		payloadPLayer.Blue, _ = strconv.Atoi(message.Text)
		payloadUpdated, _ := json.Marshal(payloadPLayer)

		setPlayerState(&player, PlayerState{
			Function: "test-multi-state",
			Stage:    4,
			Payload:  string(payloadUpdated),
		})

		msg := services.NewMessage(message.Chat.ID, "Finish?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("YES!"),
			),
		)
		services.SendMessage(msg)
	case 4:
		setPlayerState(&player, PlayerState{})

		msg := services.NewMessage(message.Chat.ID, "End")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		services.SendMessage(msg)
	}
}

// EsterEgg for debug
func theAnswerIs(update tgbotapi.Update) {
	log.Println(42)
}
