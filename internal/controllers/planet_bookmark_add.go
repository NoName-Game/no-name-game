package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

// ====================================
// PlanetBookmarkAddController
// ====================================
type PlanetBookmarkAddController struct {
	Controller
}

func (c *PlanetBookmarkAddController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.planet.bookmark.add",
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
func (c *PlanetBookmarkAddController) Handle(player *pb.Player, update tgbotapi.Update) {
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
	if _, err = config.App.Server.Connection.AddPlanetBookmark(helpers.NewContext(1), &pb.AddPlanetBookmarkRequest{
		PlanetID: playerPosition.GetID(),
		PlayerID: c.Player.GetID(),
	}); err != nil {
		if strings.Contains(err.Error(), "bookmarks limit reached") {
			// Raggiunto limite bookmark
			msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "planet.bookmark.error"))
			msg.ParseMode = tgbotapi.ModeHTML

			if _, err := helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		}
	}

	msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "planet.bookmark.add_ok"))
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
func (c *PlanetBookmarkAddController) Validator() bool {
	return false
}

// ====================================
// Stage
// ====================================
func (c *PlanetBookmarkAddController) Stage() {
	//
}
