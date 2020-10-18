package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerInventoryController
// ====================================
type PlayerInventoryController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerInventoryController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
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
		},
	}) {
		return
	}

	msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "inventory.intro"))
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.inventory.resources")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.inventory.items")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
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
