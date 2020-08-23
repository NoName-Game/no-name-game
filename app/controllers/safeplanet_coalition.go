package controllers

import (
	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Coalition
// ====================================
type SafePlanetCoalitionController struct {
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetCoalitionController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	var err error
	c.Player = player
	c.Update = update
	c.Configuration.Controller = "route.safeplanet.coalition"

	// Se tutto ok imposto e setto il nuovo stato in cache
	helpers.SetCacheState(c.Player.ID, c.Configuration.Controller)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &MenuController{}) {
			return
		}
	}

	msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "safeplanet.coalition.info"))
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.mission")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.titan")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.research")),
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

func (c *SafePlanetCoalitionController) Validator() {
	//
}

func (c *SafePlanetCoalitionController) Stage() {
	//
}
