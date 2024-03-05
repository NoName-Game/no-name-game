package controllers

import (
	"fmt"

	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerEmblemsController
// ====================================
type PlayerEmblemsController struct {
	Controller
	Payload struct {
		AchievementCategoryID uint32
	}
}

func (c *PlayerEmblemsController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.emblems",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerController{},
				FromStage: 0,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu", "route.breaker.back"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *PlayerEmblemsController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error
	// Verifico se è impossibile inizializzare
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// recupero tutti gli emblemi del player
	var rGetPlayerEmblems *pb.GetPlayerEmblemsResponse
	if rGetPlayerEmblems, err = config.App.Server.Connection.GetPlayerEmblems(helpers.NewContext(1), &pb.GetPlayerEmblemsRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Costruisco il messaggio
	var text string
	text = helpers.Trans(c.Player.Language.Slug, "emblem.intro")
	if len(rGetPlayerEmblems.GetEmblems()) > 0 {
		for _, emblem := range rGetPlayerEmblems.GetEmblems() {
			text += fmt.Sprintf("- %s ✅\n", helpers.Trans(c.Player.Language.Slug, "emblem."+emblem.Slug))
		}
	} else {
		text += helpers.Trans(c.Player.Language.Slug, "emblem.none") + "\n"
	}
	text += helpers.Trans(c.Player.Language.Slug, "emblem.footer")

	msg := helpers.NewMessage(c.ChatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}

	// Completo progressione
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *PlayerEmblemsController) Validator() (hasErrors bool) {
	return false
}

// ====================================
// Stage
// ====================================
func (c *PlayerEmblemsController) Stage() {

}
