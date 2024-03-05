package controllers

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"
)

// ====================================
// SafePlanetProtectorsChangeNameController
// ====================================
type SafePlanetProtectorsChangeNameController struct {
	Payload struct {
		Name string
	}
	Controller
}

func (c *SafePlanetProtectorsChangeNameController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.protectors.change_name",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetProtectorsController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
				2: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetProtectorsChangeNameController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
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
func (c *SafePlanetProtectorsChangeNameController) Validator() bool {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il nome della gida scelta è già stato scelto
	// ##################################################################################################
	case 1:
		var err error
		var rCheckGuildName *pb.CheckGuildNameResponse
		if rCheckGuildName, err = config.App.Server.Connection.CheckGuildName(helpers.NewContext(1), &pb.CheckGuildNameRequest{
			Name: c.Update.Message.Text,
		}); err != nil {
			c.Logger.Panic(err)
		}

		if !rCheckGuildName.GetGuildNameFree() || c.Update.Message.Text == "" {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.protectors_name_already_taken")
			return true
		}

		c.Payload.Name = c.Update.Message.Text
	// ##################################################################################################
	// Verifico conferma cambio gilda
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
func (c *SafePlanetProtectorsChangeNameController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player di inserire il nuovo della gilda
	// ##################################################################################################
	case 0:
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.change_name_start"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1
	// ##################################################################################################
	// Chiedo al player di confermare
	// ##################################################################################################
	case 1:
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.change_name_confirm", c.Payload.Name))
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
	// Creo la gilda
	// ##################################################################################################
	case 2:
		// Recupero gilda corrente
		var rGetPlayerGuild *pb.GetPlayerGuildResponse
		if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		_, err = config.App.Server.Connection.ChangeGuildName(helpers.NewContext(1), &pb.ChangeGuildNameRequest{
			GuildID: rGetPlayerGuild.GetGuild().GetID(),
			Name:    c.Payload.Name,
		})

		if err != nil && strings.Contains(err.Error(), "not enough guild points") {
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.change_name_not_enough_points"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		} else if err != nil {
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.change_name_completed_ko"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		}

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.change_name_completed_ok"))
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
