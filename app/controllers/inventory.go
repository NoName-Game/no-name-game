package controllers

import (
	pb "bitbucket.org/no-name-game/nn-grpc/rpc"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Inventory
// ====================================
type InventoryController struct {
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *InventoryController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	var err error
	c.Player = player
	c.Update = update
	c.Controller = "route.inventory"

	// Se tutto ok imposto e setto il nuovo stato su redis
	_ = helpers.SetRedisState(*c.Player, c.Controller)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &PlayerController{}) {
			return
		}
	}

	msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "inventory.intro"))
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory.recap")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory.items")),
			// tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory.destroy")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.back")),
		),
	)

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}

func (c *InventoryController) Validator() {
	//
}

func (c *InventoryController) Stage() {
	//
}
