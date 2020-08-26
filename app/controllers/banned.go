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
	c.Configuration.Controller = "route.banned"

	// Se tutto ok imposto e setto il nuovo stato in cache
	helpers.SetCacheState(c.Player.ID, c.Configuration.Controller)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if c.BackTo(0, &MenuController{}) {
		return
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
