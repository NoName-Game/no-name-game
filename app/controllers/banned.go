package controllers

import (
	pb "bitbucket.org/no-name-game/nn-grpc/rpc"
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
func (c *BannedController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	var err error
	c.Player = player
	c.Update = update
	c.Controller = "route.banned"

	// Se tutto ok imposto e setto il nuovo stato su redis
	_ = helpers.SetRedisState(*c.Player, c.Controller)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &MenuController{}) {
			return
		}
	}

	msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "banned.message"))
	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}

func (c *BannedController) Validator() {
	//
}

func (c *BannedController) Stage() {
	//
}
