package app

import (
	"encoding/json"
	"log"
	"strconv"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

//====================================
//====================================
//			TEST / EXAMPLES
//====================================
//====================================

// Back delete only redis state, but not delete state stored in DB.
func back(update tgbotapi.Update) {
	delRedisState(player)

	message := update.Message
	msg := services.NewMessage(message.Chat.ID, "LOG: Deleted only redis state without completion.")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	services.SendMessage(msg)
}

// Clears - Delete redist state and remove row from DB.
func clears(update tgbotapi.Update) {
	deleteRedisAndDbState(player)

	message := update.Message
	msg := services.NewMessage(message.Chat.ID, "LOG: Deleted redis/DB row.")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	services.SendMessage(msg)
}

// Only for testing multi-stage
func testMultiStage(update tgbotapi.Update) {
	//====================================
	// Init Func!
	//====================================
	message := update.Message
	routeName := "test-multi-stage"
	state := startAndCreateState(routeName)

	//====================================
	// Validator
	//====================================
	validationFlag := true
	validationMessage := "Wrong input, please repeat or exit."
	switch state.Stage {
	case 0:
		if message.Text == "Go to stage 1" {
			state.Stage = 1
			state.update()
			validationFlag = false
		}
	case 1:
		if message.Text == "YES!" {
			state.Stage = 2
			state.update()
			validationFlag = false
		}
	}

	if true == validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			services.SendMessage(validatorMsg)
		}
	}

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		msg := services.NewMessage(message.Chat.ID, "This is stage 0.")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Go to stage 1"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
				tgbotapi.NewKeyboardButton("clears"),
			),
		)
		services.SendMessage(msg)
	case 1:
		msg := services.NewMessage(message.Chat.ID, "Finish?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("YES!"),
				tgbotapi.NewKeyboardButton("Wrong answare!"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
			),
		)
		services.SendMessage(msg)
	case 2:
		//====================================
		// IMPORTANT!
		//====================================
		finishAndCompleteState(state)
		//====================================

		msg := services.NewMessage(message.Chat.ID, "Completed! :)")
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
	//====================================
	// Init Func!
	//====================================
	type payloadStruct struct {
		Red   int
		Green int
		Blue  int
	}

	message := update.Message
	routeName := "test-multi-state"
	state := startAndCreateState(routeName)
	var payload payloadStruct
	unmarshalPayload(state.Payload, &payload)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := "Wrong input, please repeat or exit."
	switch state.Stage {
	case 0:
		input, _ := strconv.Atoi(message.Text)
		if input >= 1 && input <= 100 {
			state.Stage = 1
			state.update()
			validationFlag = true
		}
	case 1:
		input, _ := strconv.Atoi(message.Text)
		if input >= 1 && input <= 100 {
			state.Stage = 2
			state.update()
			validationFlag = true
		}
	case 2:
		input, _ := strconv.Atoi(message.Text)
		if input >= 1 && input <= 100 {
			state.Stage = 3
			state.update()
			validationFlag = true
		}
	case 3:
		if message.Text == "YES!" {
			state.Stage = 4
			state.update()
			validationFlag = true
		}
	}

	if false == validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			services.SendMessage(validatorMsg)
		}
	}

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(payloadStruct{})
		state.Payload = string(payloadUpdated)
		state.update()

		msg := services.NewMessage(message.Chat.ID, "State setted, How much of R?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("50"),
				tgbotapi.NewKeyboardButton("100"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
				tgbotapi.NewKeyboardButton("clears"),
			),
		)
		services.SendMessage(msg)

	case 1:
		// If is valid input
		if true == validationFlag {
			//R
			payload.Red, _ = strconv.Atoi(message.Text)
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.update()
		}

		msg := services.NewMessage(message.Chat.ID, "Stage 2 setted, How much of G?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("50"),
				tgbotapi.NewKeyboardButton("100"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
				tgbotapi.NewKeyboardButton("clears"),
			),
		)
		services.SendMessage(msg)
	case 2:
		// If is valid input
		if true == validationFlag {
			//G
			payload.Green, _ = strconv.Atoi(message.Text)
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.update()
		}

		msg := services.NewMessage(message.Chat.ID, "Stage 2 setted, How much of B?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("50"),
				tgbotapi.NewKeyboardButton("100"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
				tgbotapi.NewKeyboardButton("clears"),
			),
		)
		services.SendMessage(msg)
	case 3:
		// If is valid input
		if true == validationFlag {
			//B
			payload.Blue, _ = strconv.Atoi(message.Text)
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.update()
		}

		msg := services.NewMessage(message.Chat.ID, "Finish?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("YES!"),
				tgbotapi.NewKeyboardButton("Wrong answare!"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
				tgbotapi.NewKeyboardButton("clears"),
			),
		)
		services.SendMessage(msg)
	case 4:
		//====================================
		// IMPORTANT!
		//====================================
		finishAndCompleteState(state)
		//====================================

		msg := services.NewMessage(message.Chat.ID, "Completed :)")
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
