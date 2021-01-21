package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerTeamController
// ====================================
type PlayerTeamController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerTeamController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.team",
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

	// Recupero team player
	var rGetTeamDetails *pb.GetTeamDetailsResponse
	rGetTeamDetails, _ = config.App.Server.Connection.GetTeamDetails(helpers.NewContext(1), &pb.GetTeamDetailsRequest{
		PlayerID: c.Player.ID,
	})

	// Se il player si trova in un team recupero i dettagli
	if rGetTeamDetails.GetInTeam() {
		// Recap owner
		var ownerRecap string
		ownerRecap = fmt.Sprintf("Owner: %s \n\n", rGetTeamDetails.GetOwner().GetUsername())

		// Ciclio utenti nel team
		var playerRecap string
		for _, player := range rGetTeamDetails.GetPlayers() {
			playerRecap += fmt.Sprintf("- %s \n", player.GetUsername())
		}

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, fmt.Sprintf("%s %s", ownerRecap, playerRecap))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.team.leave")),
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.team.add_player")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}
		return
	}

	// Il Player non è in un team
	msg := helpers.NewMessage(c.Update.Message.Chat.ID, fmt.Sprintf("Non in team"))
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.team.create")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
		),
	)

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *PlayerTeamController) Validator() bool {
	return false
}

func (c *PlayerTeamController) Stage() {
	//
}
