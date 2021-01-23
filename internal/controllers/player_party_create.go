package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerPartyCreateController
// ====================================
type PlayerPartyCreateController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerPartyCreateController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.party.create",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerPartyController{},
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
func (c *PlayerPartyCreateController) Validator() bool {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il player appartiene già ad un party
	// ##################################################################################################
	case 0:
		// Recupero party player
		var rGetPartyDetails *pb.GetPartyDetailsResponse
		rGetPartyDetails, _ = config.App.Server.Connection.GetPartyDetails(helpers.NewContext(1), &pb.GetPartyDetailsRequest{
			PlayerID: c.Player.ID,
		})

		// Se il player si trova in un party recupero i dettagli
		if rGetPartyDetails.GetInParty() {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "player.party.already_in_party")
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *PlayerPartyCreateController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Creo party player
	// ##################################################################################################
	case 0:
		if _, err = config.App.Server.Connection.CreateParty(helpers.NewContext(1), &pb.CreatePartyRequest{
			OwnerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.party.create.created"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Completed = true
	}
}
