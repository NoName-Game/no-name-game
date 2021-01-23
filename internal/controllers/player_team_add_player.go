package controllers

import (
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerTeamAddPlayerController
// ====================================
type PlayerTeamAddPlayerController struct {
	Payload struct {
		Username string
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerTeamAddPlayerController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.team.add_player",
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
func (c *PlayerTeamAddPlayerController) Validator() bool {
	// Verifico sempre che il player sia owner del team, se no non può eseguire questi comandi
	var rGetTeamDetails *pb.GetTeamDetailsResponse
	rGetTeamDetails, _ = config.App.Server.Connection.GetTeamDetails(helpers.NewContext(1), &pb.GetTeamDetailsRequest{
		PlayerID: c.Player.ID,
	})

	if rGetTeamDetails.GetOwner().GetID() != c.Player.ID {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "player.team.not_leader")
		return true
	}

	// Verifico se non è già stato raggiunto il limite di player in team
	if rGetTeamDetails.NPlayers >= 3 {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "player.team.team_limit_reached")
		return true
	}

	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il player scelto esiste
	// ##################################################################################################
	case 1:
		var rGetPlayerByUsername *pb.GetPlayerByUsernameResponse
		rGetPlayerByUsername, _ = config.App.Server.Connection.GetPlayerByUsername(helpers.NewContext(1), &pb.GetPlayerByUsernameRequest{
			Username: c.Update.Message.Text,
		})

		if rGetPlayerByUsername.GetPlayer().GetID() <= 0 || c.Update.Message.Text == "" {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.player_name_not_exists")
			return true
		}

		c.Payload.Username = c.Update.Message.Text
	// ##################################################################################################
	// Verifico Conferma
	// ##################################################################################################
	case 2:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *PlayerTeamAddPlayerController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player di indicare quale player vuole aggiungere alla suo team
	// ##################################################################################################
	case 0:
		// Aggiungo torna al menu
		var protectorsKeyboard [][]tgbotapi.KeyboardButton
		protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		))

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.team.add.add_player_start"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       protectorsKeyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1
	// ##################################################################################################
	// Chiedo Conferma al player
	// ##################################################################################################
	case 1:
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.team.add.add_player_confirm", c.Payload.Username))
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
	// Salvo e associo player alla gilda
	// ##################################################################################################
	case 2:
		// Recupero team player
		var rGetTeamDetails *pb.GetTeamDetailsResponse
		rGetTeamDetails, _ = config.App.Server.Connection.GetTeamDetails(helpers.NewContext(1), &pb.GetTeamDetailsRequest{
			PlayerID: c.Player.ID,
		})

		_, err = config.App.Server.Connection.AddPlayerToTeam(helpers.NewContext(1), &pb.AddPlayerToTeamRequest{
			PlayerUsername: c.Payload.Username,
			TeamID:         rGetTeamDetails.TeamID,
		})

		if err != nil && strings.Contains(err.Error(), "player already in one team") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "player.team.adding_player_already_in_one_team"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		}

		// Invio messaggio al player aggiunto
		var rGetPlayerByUsername *pb.GetPlayerByUsernameResponse
		rGetPlayerByUsername, _ = config.App.Server.Connection.GetPlayerByUsername(helpers.NewContext(1), &pb.GetPlayerByUsernameRequest{
			Username: c.Payload.Username,
		})

		msgToPlayerAdded := helpers.NewMessage(rGetPlayerByUsername.GetPlayer().GetChatID(), helpers.Trans(
			rGetPlayerByUsername.GetPlayer().GetLanguage().GetSlug(),
			"player.team.add.add_player_confirm_to_player", c.Player.GetUsername(),
		))
		msgToPlayerAdded.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msgToPlayerAdded); err != nil {
			c.Logger.Panic(err)
		}

		// Ritorno conferma
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.team.add.completed_ok"))
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
