package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerTeamCreateController
// ====================================
type PlayerTeamCreateController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerTeamCreateController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.team.create",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerTeamController{},
				FromStage: 0,
			},
		},
	}) {
		return
	}

	// Validate
	if c.Validator() {
		c.Validate()
		return
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(nil)
}

// ====================================
// Validator
// ====================================
func (c *PlayerTeamCreateController) Validator() bool {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il player appartiene già ad un team
	// ##################################################################################################
	case 0:
		// Recupero team player
		var rGetTeamDetails *pb.GetTeamDetailsResponse
		rGetTeamDetails, _ = config.App.Server.Connection.GetTeamDetails(helpers.NewContext(1), &pb.GetTeamDetailsRequest{
			PlayerID: c.Player.ID,
		})

		// Se il player si trova in un team recupero i dettagli
		if rGetTeamDetails.GetInTeam() {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "player.team.already_in_team")
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *PlayerTeamCreateController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Creo team player
	// ##################################################################################################
	case 0:
		if _, err = config.App.Server.Connection.CreateTeam(helpers.NewContext(1), &pb.CreateTeamRequest{
			OwnerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.team.create.created"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Completed = true
	}
}
