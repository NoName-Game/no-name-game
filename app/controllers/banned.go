package controllers

import (
	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// BannedController
// ====================================
type BannedController struct {
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *BannedController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error
	c.Player = player
	c.Update = update

	// Init Controller
	if !c.InitController(ControllerConfiguration{
		Controller: "route.banned",
		ControllerBack: ControllerBack{
			To:        &MenuController{},
			FromStage: 0,
		},
	}) {
		return
	}

	msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "banned.message"))
	if _, err = services.SendMessage(msg); err != nil {
		panic(err)
	}
}

func (c *BannedController) Validator() {
	//
}

func (c *BannedController) Stage() {
	//
}
