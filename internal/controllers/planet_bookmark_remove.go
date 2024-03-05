package controllers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"
)

// ====================================
// PlanetBookmarkRemoveController
// ====================================
type PlanetBookmarkRemoveController struct {
	Controller
}

func (c *PlanetBookmarkRemoveController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.planet.bookmark.remove",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlanetController{},
				FromStage: 0,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *PlanetBookmarkRemoveController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Recupero posizione player corrente
	var playerPosition *pb.Planet
	if playerPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
		c.Logger.Panic(err)
	}

	// Aggiungo pianeta ai preferiti
	_, _ = config.App.Server.Connection.RemovePlanetBookmark(helpers.NewContext(1), &pb.RemovePlanetBookmarkRequest{
		PlanetID: playerPosition.GetID(),
		PlayerID: c.Player.GetID(),
	})

	msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "planet.bookmark.remove_ok"))
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)

	if _, err := helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *PlanetBookmarkRemoveController) Validator() bool {
	return false
}

// ====================================
// Stage
// ====================================
func (c *PlanetBookmarkRemoveController) Stage() {
	//
}
