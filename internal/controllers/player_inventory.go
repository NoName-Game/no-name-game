package controllers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/internal/helpers"
)

// ====================================
// PlayerInventoryController
// ====================================
type PlayerInventoryController struct {
	Controller
}

func (c *PlayerInventoryController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.inventory",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerController{},
				FromStage: 0,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
			AllowedControllers: []string{
				"route.player.inventory.resources",
				"route.player.inventory.items",
				"route.player.inventory.packs",
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *PlayerInventoryController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "inventory.intro"))
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.inventory.resources")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.inventory.items")),
		),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.inventory.packs"))),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)

	if _, err := helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *PlayerInventoryController) Validator() bool {
	return false
}

func (c *PlayerInventoryController) Stage() {
	//
}
