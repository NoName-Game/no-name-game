package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlanetBookmarkAddController
// ====================================
type PlanetBookmarkAddController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlanetBookmarkAddController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(Controller{
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
		},
	}) {
		return
	}

	// Recupero posizione player corrente
	var playerPosition *pb.Planet
	if playerPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
		c.Logger.Panic(err)
	}

	// Aggiungo pianeta ai preferiti
	_, _ = config.App.Server.Connection.AddPlanetBookmark(helpers.NewContext(1), &pb.AddPlanetBookmarkRequest{
		PlanetID: playerPosition.GetID(),
		PlayerID: c.Player.GetID(),
	})

	msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "planet.bookmark.add_ok"))
	msg.ParseMode = tgbotapi.ModeMarkdown
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
