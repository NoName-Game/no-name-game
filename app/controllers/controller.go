package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BaseController struct {
	Update     tgbotapi.Update
	Message    *tgbotapi.Message
	RouteName  string
	Validation struct {
		HasErrors bool
		Message   string
	}
	State nnsdk.PlayerState
}
