package controllers

import (
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetProtectorsCreateController
// ====================================
type SafePlanetProtectorsCreateController struct {
	Payload struct {
		Name          string
		Accessibility bool // Pubblico o privato
	}
	Controller
}

func (c *SafePlanetProtectorsCreateController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.protectors.create",
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
				2: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetProtectorsCreateController) Handle(player *pb.Player, update tgbotapi.Update) {
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
func (c *SafePlanetProtectorsCreateController) Validator() bool {
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
	// Verifico tipologia di gilda pubblica o privata
	// ##################################################################################################
	case 2:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.create_accessibility.public") {
			c.Payload.Accessibility = false
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.create_accessibility.private") {
			c.Payload.Accessibility = true
			return false
		}

		return true
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetProtectorsCreateController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player di inserire il nome della gilda
	// ##################################################################################################
	case 0:
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.create_start"))
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
	// Chiedo al player di inserire di decidere se al gilda deve esser pubblica i privata
	// ##################################################################################################
	case 1:
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.create_accessibility"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.create_accessibility.public")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.create_accessibility.private")),
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
		_, err = config.App.Server.Connection.CreateGuild(helpers.NewContext(1), &pb.CreateGuildRequest{
			GuildName: c.Payload.Name,
			OwnerID:   c.Player.ID,
			GuildType: c.Payload.Accessibility,
		})

		if err != nil && strings.Contains(err.Error(), "player already in one guild") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.player_already_in_one_protectors"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		}

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.creation_completed_ok"))
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
