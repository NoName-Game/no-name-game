package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerTeamLeaveController
// ====================================
type PlayerTeamLeaveController struct {
	Payload struct {
		Username string
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerTeamLeaveController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.team.leave",
			Payload:    &c.Payload,
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
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *PlayerTeamLeaveController) Validator() bool {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico Conferma
	// ##################################################################################################
	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *PlayerTeamLeaveController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo Conferma al player
	// ##################################################################################################
	case 0:
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.team.leave.confirm"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 2
	// ##################################################################################################
	// Levo player dal team
	// ##################################################################################################
	case 2:
		// Recupero team player
		var rGetTeamDetails *pb.GetTeamDetailsResponse
		rGetTeamDetails, _ = config.App.Server.Connection.GetTeamDetails(helpers.NewContext(1), &pb.GetTeamDetailsRequest{
			PlayerID: c.Player.ID,
		})

		if _, err = config.App.Server.Connection.RemovePlayerFromTeam(helpers.NewContext(1), &pb.RemovePlayerFromTeamRequest{
			PlayerID: c.Player.ID,
			TeamID:   rGetTeamDetails.GetTeamID(),
		}); err != nil {
			c.Logger.Warning(err)

			// Potrebbero esserci stati degli errori generici
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "player.team.leave.completed_ko"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		}

		// Se il player è l'owner del team devo mandare il mesaggio anche a tutti gli altri
		if rGetTeamDetails.GetOwner().GetID() == c.Player.GetID() {
			for _, player := range rGetTeamDetails.GetPlayers() {
				// Lo mando a tutti tranne che a me stesso
				if player.GetID() != c.Player.GetID() {
					var rGetPlayerByUsername *pb.GetPlayerByUsernameResponse
					rGetPlayerByUsername, _ = config.App.Server.Connection.GetPlayerByUsername(helpers.NewContext(1), &pb.GetPlayerByUsernameRequest{
						Username: player.GetUsername(),
					})

					msgToPlayerRemoved := helpers.NewMessage(rGetPlayerByUsername.GetPlayer().GetChatID(), helpers.Trans(
						rGetPlayerByUsername.GetPlayer().GetLanguage().GetSlug(),
						"player.team.remove.remove_player_confirm_to_player", c.Player.GetUsername(),
					))
					msgToPlayerRemoved.ParseMode = tgbotapi.ModeMarkdown
					if _, err = helpers.SendMessage(msgToPlayerRemoved); err != nil {
						c.Logger.Panic(err)
					}
				}
			}
		}

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.team.leave.completed_ok"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Completed = true
	}
}
