package controllers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/internal/helpers"
)

// ====================================
// Coalition
// ====================================
type SafePlanetCoalitionController struct {
	Controller
}

func (c *SafePlanetCoalitionController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
			AllowedControllers: []string{
				"route.safeplanet.coalition.daily_reward",
				"route.safeplanet.coalition.mission",
				"route.safeplanet.coalition.titan",
				"route.safeplanet.coalition.expansion",
				"route.safeplanet.coalition.research",
				"route.safeplanet.coalition.protectors",
				"route.safeplanet.coalition.statistics",
				"route.info",
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetCoalitionController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "safeplanet.coalition.info"))
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.daily_reward")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.mission")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.titan")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.expansion")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.research")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.statistics")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.info")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *SafePlanetCoalitionController) Validator() bool {
	return false
}

func (c *SafePlanetCoalitionController) Stage() {
	//
}
