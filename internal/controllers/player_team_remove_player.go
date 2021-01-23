package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerTeamRemovePlayerController
// ====================================
type PlayerTeamRemovePlayerController struct {
	Payload struct {
		Username string
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerTeamRemovePlayerController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.team.remove_player",
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
func (c *PlayerTeamRemovePlayerController) Validator() bool {
	// Verifico sempre che il player sia owner del team, se no non può eseguire questi comandi
	var rGetTeamDetails *pb.GetTeamDetailsResponse
	rGetTeamDetails, _ = config.App.Server.Connection.GetTeamDetails(helpers.NewContext(1), &pb.GetTeamDetailsRequest{
		PlayerID: c.Player.ID,
	})

	if rGetTeamDetails.GetOwner().GetID() != c.Player.ID {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "player.team.not_leader")
		return true
	}

	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il player scelto esisteo
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
func (c *PlayerTeamRemovePlayerController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player di indicare quale player vuole aggiungere alla sua gilda
	// ##################################################################################################
	case 0:
		// Verifico/Recupero Gilda player
		var err error
		var rGetTeamDetails *pb.GetTeamDetailsResponse
		if rGetTeamDetails, err = config.App.Server.Connection.GetTeamDetails(helpers.NewContext(1), &pb.GetTeamDetailsRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		var protectorsKeyboard [][]tgbotapi.KeyboardButton
		for _, player := range rGetTeamDetails.GetPlayers() {
			if player.GetID() != c.Player.GetID() {
				protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(player.GetUsername()),
				))
			}
		}

		// Aggiungo torna al menu
		protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		))

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.team.remove.remove_player_start"))
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
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.team.remove.remove_player_confirm", c.Payload.Username))
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
	// Salvo e associo player al team
	// ##################################################################################################
	case 2:
		// Recupero team player
		var rGetTeamDetails *pb.GetTeamDetailsResponse
		rGetTeamDetails, _ = config.App.Server.Connection.GetTeamDetails(helpers.NewContext(1), &pb.GetTeamDetailsRequest{
			PlayerID: c.Player.ID,
		})

		if _, err = config.App.Server.Connection.RemovePlayerToTeam(helpers.NewContext(1), &pb.RemovePlayerToTeamRequest{
			PlayerUsername: c.Payload.Username,
			TeamID:         rGetTeamDetails.GetTeamID(),
		}); err != nil {
			c.Logger.Warning(err)

			// Potrebbero esserci stati degli errori generici
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "player.team.remove.remove_completed_ko"),
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

		msgToPlayerRemoved := helpers.NewMessage(rGetPlayerByUsername.GetPlayer().GetChatID(), helpers.Trans(
			rGetPlayerByUsername.GetPlayer().GetLanguage().GetSlug(),
			"player.team.remove.remove_player_confirm_to_player", c.Player.GetUsername(),
		))
		msgToPlayerRemoved.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msgToPlayerRemoved); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "player.team.remove.remove_completed_ok"))
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
