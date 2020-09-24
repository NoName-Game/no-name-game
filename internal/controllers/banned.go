package controllers

import (
	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// BannedController
// ====================================
type BannedController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *BannedController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
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
	}) {
		return
	}

	msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "banned.message"))
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
