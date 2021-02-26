package controllers

import (
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerPartyAddPlayerController
// ====================================
type PlayerPartyAddPlayerController struct {
	Payload struct {
		Username string
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerPartyAddPlayerController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.party.add_player",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerPartyController{},
				FromStage: 0,
			},
			BreakerPerStage: map[int32][]string{
				1: {"route.breaker.menu"},
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
func (c *PlayerPartyAddPlayerController) Validator() bool {
	// Verifico sempre che il player sia owner del party, se no non può eseguire questi comandi
	var rGetPartyDetails *pb.GetPartyDetailsResponse
	rGetPartyDetails, _ = config.App.Server.Connection.GetPartyDetails(helpers.NewContext(1), &pb.GetPartyDetailsRequest{
		PlayerID: c.Player.ID,
	})

	if rGetPartyDetails.GetOwner().GetID() != c.Player.ID {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "player.party.not_leader")
		return true
	}

	// Verifico se non è già stato raggiunto il limite di player in party
	if rGetPartyDetails.NPlayers >= 3 {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "player.party.party_limit_reached")
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
func (c *PlayerPartyAddPlayerController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player di indicare quale player vuole aggiungere alla suo party
	// ##################################################################################################
	case 0:
		// Aggiungo torna al menu
		var protectorsKeyboard [][]tgbotapi.KeyboardButton
		protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "player.party.add.add_player_start"))
		msg.ParseMode = tgbotapi.ModeHTML
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
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "player.party.add.add_player_confirm", c.Payload.Username))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
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
		// Recupero party player
		var rGetPartyDetails *pb.GetPartyDetailsResponse
		rGetPartyDetails, _ = config.App.Server.Connection.GetPartyDetails(helpers.NewContext(1), &pb.GetPartyDetailsRequest{
			PlayerID: c.Player.ID,
		})

		_, err = config.App.Server.Connection.AddPlayerToParty(helpers.NewContext(1), &pb.AddPlayerToPartyRequest{
			PlayerUsername: c.Payload.Username,
			PartyID:        rGetPartyDetails.PartyID,
		})

		if err != nil && strings.Contains(err.Error(), "player already in one party") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "player.party.adding_player_already_in_one_party"),
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
			"player.party.add.add_player_confirm_to_player", c.Player.GetUsername(),
		))
		msgToPlayerAdded.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msgToPlayerAdded); err != nil {
			c.Logger.Panic(err)
		}

		// Ritorno conferma
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "player.party.add.completed_ok"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Completed = true
	}
}
