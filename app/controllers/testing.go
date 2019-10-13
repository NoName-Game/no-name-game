package controllers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type TestingController struct {
	RouteName  string
	Validation bool
	Update     tgbotapi.Update
	Message    *tgbotapi.Message
	Payload    struct{}
}

//====================================
// Handle
//====================================
func (t TestingController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	t.RouteName = "route.testing.multiStageRevelution"
	t.Message = update.Message

	// Check current state for this routes
	state, isNewState := helpers.CheckState(t.RouteName, t.Payload, helpers.Player)

	// It's first message
	if isNewState {
		t.Stage(state)
		return
	}

	// Go to validator
	t.Validation, state = t.Validator(state)
	if !t.Validation {
		state, _ = providers.UpdatePlayerState(state)
		t.Stage(state)
	}
	return
}

//====================================
// Validator
//====================================
func (t TestingController) Validator(state nnsdk.PlayerState) (hasErrors bool, newState nnsdk.PlayerState) {
	switch state.Stage {
	case 0:
		if t.Message.Text == "Go to stage 1" {
			state.Stage = 1
			return false, state
		}
	case 1:
		if t.Message.Text == "YES!" {
			state.Stage = 2
			return false, state
		}
	}

	// Validator goes errors
	validatorMsg := services.NewMessage(t.Message.Chat.ID, "Wrong input, please repeat or exit.")
	services.SendMessage(validatorMsg)

	return true, state
}

//====================================
// Stage
//====================================
func (t TestingController) Stage(state nnsdk.PlayerState) {
	switch state.Stage {
	case 0:
		msg := services.NewMessage(t.Message.Chat.ID, "This is stage 0.")
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
		msg := services.NewMessage(t.Message.Chat.ID, "Finish?")
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
		helpers.FinishAndCompleteState(state, helpers.Player)
		//====================================

		msg := services.NewMessage(t.Message.Chat.ID, "Completed! :)")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
			),
		)
		services.SendMessage(msg)
	}
}
