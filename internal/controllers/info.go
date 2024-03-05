package controllers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/internal/helpers"
)

// ====================================
// BannedController
// ====================================
type InfoController struct {
	Controller
}

func (c *InfoController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{}
}

// ====================================
// Handle
// ====================================
func (c *InfoController) Handle(player *pb.Player, update tgbotapi.Update) {
	msg := helpers.NewMessage(player.ChatID, helpers.Trans(player.Language.Slug, "info"))
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *InfoController) Validator() bool {
	return false
}

func (c *InfoController) Stage() {
	//
}
