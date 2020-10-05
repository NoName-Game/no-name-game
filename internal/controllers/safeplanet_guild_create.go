package controllers

import (
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetGuildCreateController
// ====================================
type SafePlanetGuildCreateController struct {
	Payload struct {
		Name          string
		Accessibility bool // Pubblico o privato
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetGuildCreateController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.guild.create",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetGuildController{},
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
func (c *SafePlanetGuildCreateController) Validator() bool {
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
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.guild_name_already_taken")
			return true
		}

		c.Payload.Name = c.Update.Message.Text
	// ##################################################################################################
	// Verifico tipologia di gilda pubblica o privata
	// ##################################################################################################
	case 2:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.create_accessibility.public") {
			c.Payload.Accessibility = false
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.create_accessibility.private") {
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
func (c *SafePlanetGuildCreateController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player di inserire il nome della gilda
	// ##################################################################################################
	case 0:
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.create_start"))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
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
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.create_accessibility"))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.create_accessibility.public")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.create_accessibility.private")),
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
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.player_already_in_one_guild"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		}

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.creation_completed_ok"))
		msg.ParseMode = "markdown"
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
