package controllers

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	"bitbucket.org/no-name-game/no-name/app/commands"
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//====================================
//====================================
//			TEST / EXAMPLES
//====================================
//====================================

// TestTimedQuest - ...
func TestTimedQuest(update tgbotapi.Update, player models.Player) {
	message := update.Message
	routeName := "timed-quest"
	state := helpers.StartAndCreatePlayerState(routeName, player)

	go commands.Cron(1 * time.Minute) //Check every minute

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := "Wrong input, please repeat or exit."
	switch state.Stage {
	case 0:
		if message.Text == "ok" {
			state.Stage = 1
			state.Update()
			validationFlag = false
		}
	case 1:
		if time.Now().Before(state.FinishAt) {
			validationFlag = false
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			services.SendMessage(validatorMsg)
		}
	}

	switch state.Stage {
	case 0:
		state.Stage = 1
		state.FinishAt = time.Now().Add((time.Minute*time.Duration(10) + time.Second*time.Duration(15)))
		state.Update()
		log.Println(state.FinishAt)
		msg := services.NewMessage(message.Chat.ID, state.FinishAt.String())
		services.SendMessage(msg)
	case 1:
		if validationFlag {
			helpers.FinishAndCompleteState(state, player)
		}
	}

}

// TestMultiStage - Only for testing multi-stage
func TestMultiStage(update tgbotapi.Update, player models.Player) {
	//====================================
	// Init Func!
	//====================================
	message := update.Message
	routeName := "test-multi-stage"
	state := helpers.StartAndCreatePlayerState(routeName, player)

	//====================================
	// Validator
	//====================================
	validationFlag := true
	validationMessage := "Wrong input, please repeat or exit."
	switch state.Stage {
	case 0:
		if message.Text == "Go to stage 1" {
			state.Stage = 1
			state.Update()
			validationFlag = false
		}
	case 1:
		if message.Text == "YES!" {
			state.Stage = 2
			state.Update()
			validationFlag = false
		}
	}

	if validationFlag {
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
		helpers.FinishAndCompleteState(state, player)
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

// TestMultiState - Only for testing multi-state
func TestMultiState(update tgbotapi.Update, player models.Player) {
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
	state := helpers.StartAndCreatePlayerState(routeName, player)
	var payload payloadStruct
	helpers.UnmarshalPayload(state.Payload, &payload)

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
			state.Update()
			validationFlag = true
		}
	case 1:
		input, _ := strconv.Atoi(message.Text)
		if input >= 1 && input <= 100 {
			state.Stage = 2
			state.Update()
			validationFlag = true
		}
	case 2:
		input, _ := strconv.Atoi(message.Text)
		if input >= 1 && input <= 100 {
			state.Stage = 3
			state.Update()
			validationFlag = true
		}
	case 3:
		if message.Text == "YES!" {
			state.Stage = 4
			state.Update()
			validationFlag = true
		}
	}

	if !validationFlag {
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
		state.Update()

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
		if validationFlag {
			//R
			payload.Red, _ = strconv.Atoi(message.Text)
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Update()
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
		if validationFlag {
			//G
			payload.Green, _ = strconv.Atoi(message.Text)
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Update()
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
		if validationFlag {
			//B
			payload.Blue, _ = strconv.Atoi(message.Text)
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Update()
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
		helpers.FinishAndCompleteState(state, player)
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

// TheAnswerIs - TheAnswerIs
func TheAnswerIs(update tgbotapi.Update, player models.Player) {
	log.Println(42)
}
