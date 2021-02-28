package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// BannedController
// ====================================
type InfoController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *InfoController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.info",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
		},
	}) {
		return
	}

	msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "info"))
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
	new(MenuController).Handle(c.Player, c.Update)
}

func (c *InfoController) Validator() bool {
	return false
}

func (c *InfoController) Stage() {
	//
}
