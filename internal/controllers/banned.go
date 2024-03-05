package controllers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/internal/helpers"
)

// ====================================
// BannedController
// ====================================
type BannedController struct {
	Controller
}

func (c *BannedController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.banned",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *BannedController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "banned.message"))
	if _, err := helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *BannedController) Validator() bool {
	return false
}

func (c *BannedController) Stage() {
	//
}
