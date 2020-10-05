package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetGuildController
// ====================================
type SafePlanetGuildController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetGuildController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.guild",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 0,
			},
		},
	}) {
		return
	}

	msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "safeplanet.coalition.guild.info"))
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.guild.create")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.guild.join")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
		),
	)

	if _, err := helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *SafePlanetGuildController) Validator() bool {
	return false
}

func (c *SafePlanetGuildController) Stage() {
	//
}
