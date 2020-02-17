package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BaseController struct {
	Update tgbotapi.Update
	// Callback   tgbotapi.CallbackQuery
	Message    *tgbotapi.Message
	Controller string
	Father     uint
	Validation struct {
		HasErrors bool
		Message   string
	}
	Player nnsdk.Player
	State  nnsdk.PlayerState
}
