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
func (c *SafePlanetCoalitionController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error
	c.Player = player
	c.Update = update

	// Init Controller
	if !c.InitController(ControllerConfiguration{
		Controller: "route.safeplanet.coalition",
		ControllerBack: ControllerBack{
			To:        &MenuController{},
			FromStage: 0,
		},
	}, nil) {
		return
	}

	msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "safeplanet.coalition.info"))
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.mission")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.titan")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.expansion")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.research")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.back")),
		),
	)
	if _, err = services.SendMessage(msg); err != nil {
		panic(err)
	}
}

func (c *SafePlanetCoalitionController) Validator() {
	//
}

func (c *SafePlanetCoalitionController) Stage() {
	//
}
