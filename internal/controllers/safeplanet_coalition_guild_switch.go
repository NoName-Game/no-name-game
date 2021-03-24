package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetProtectorsAddPlayerController
// ====================================
type SafePlanetProtectorsSwitchController struct {
	Payload struct {
		Visibility bool
		GuildID    uint32
	}
	Controller
}

func (c *SafePlanetProtectorsSwitchController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.protectors.switch",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetProtectorsController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				1: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetProtectorsSwitchController) Handle(player *pb.Player, update tgbotapi.Update) {
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
func (c *SafePlanetProtectorsSwitchController) Validator() bool {
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
		if rGetPlayerGuild.GetGuild().GetOwnerID() != c.Player.ID {
			c.CurrentState.Completed = true
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.not_owner")

			return true
		}

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
func (c *SafePlanetProtectorsSwitchController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player di indicare quale player vuole aggiungere alla sua gilda
	// ##################################################################################################
	case 0:
		// Aggiungo torna al menu
		var protectorsKeyboard [][]tgbotapi.KeyboardButton
		protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		// Verifico sia fondatore
		var rGetPlayerGuild *pb.GetPlayerGuildResponse
		if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		var visibility string
		if rGetPlayerGuild.GetGuild().GetGuildType() {
			visibility = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.create_accessibility.private")
		} else {
			visibility = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.create_accessibility.public")
		}

		c.Payload.Visibility = rGetPlayerGuild.GetGuild().GetGuildType()
		c.Payload.GuildID = rGetPlayerGuild.GetGuild().GetID()

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.switch", visibility))
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
	// Salvo e associo player alla gilda
	// ##################################################################################################
	case 1:
		// Recuero gilda corrente
		if _, err = config.App.Server.Connection.ChangeGuildVisibility(helpers.NewContext(1), &pb.ChangeVisibilityGuildRequest{
			Visibility: !c.Payload.Visibility,
			GuildID:    c.Payload.GuildID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.switch_conferm"))
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
