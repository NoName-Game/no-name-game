package controllers

import (
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetProtectorsAddPlayerController
// ====================================
type SafePlanetProtectorsAddPlayerController struct {
	Payload struct {
		Username string
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetProtectorsAddPlayerController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.protectors.add_player",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetProtectorsController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
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
func (c *SafePlanetProtectorsAddPlayerController) Validator() bool {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		// Verifico sia fondatore
		var rGetPlayerGuild *pb.GetPlayerGuildResponse
		if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}
		if rGetPlayerGuild.GetGuild().GetOwnerID() == c.Player.ID {
			c.CurrentState.Completed = true
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.not_owner")

			return true
		}

	// ##################################################################################################
	// Verifico se il player scelto esisteo
	// ##################################################################################################
	case 1:
		var rGetPlayerByUsername *pb.GetPlayerByUsernameResponse
		rGetPlayerByUsername, _ = config.App.Server.Connection.GetPlayerByUsername(helpers.NewContext(1), &pb.GetPlayerByUsernameRequest{
			Username: c.Update.Message.Text,
		})

		if rGetPlayerByUsername.GetPlayer().GetID() <= 0 || c.Update.Message.Text == "" {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.protectors_player_not_exists")
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
func (c *SafePlanetProtectorsAddPlayerController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player di indicare quale player vuole aggiungere alla sua gilda
	// ##################################################################################################
	case 0:
		// Aggiungo torna al menu
		var protectorsKeyboard [][]tgbotapi.KeyboardButton
		protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.add_player_start"))
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
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.add_player_confirm", c.Payload.Username))
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
		// Recuero gilda corrente
		var rGetPlayerGuild *pb.GetPlayerGuildResponse
		if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}
		_, err = config.App.Server.Connection.AddPlayerToGuild(helpers.NewContext(1), &pb.AddPlayerToGuildRequest{
			PlayerUsername: c.Payload.Username,
			GuildID:        rGetPlayerGuild.GetGuild().GetID(),
		})

		if err != nil && strings.Contains(err.Error(), "player already in one guild") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.adding_player_already_in_one_protectors"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		}

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.add_completed_ok"))
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
