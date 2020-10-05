package controllers

import (
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetGuildJoinController
// ====================================
type SafePlanetGuildJoinController struct {
	Payload struct {
		Name string
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetGuildJoinController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.guild.join",
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
func (c *SafePlanetGuildJoinController) Validator() bool {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il nome della gida scelta esiste
	// ##################################################################################################
	case 1:
		var err error
		var rCheckGuildName *pb.CheckGuildNameResponse
		if rCheckGuildName, err = config.App.Server.Connection.CheckGuildName(helpers.NewContext(1), &pb.CheckGuildNameRequest{
			Name: c.Update.Message.Text,
		}); err != nil {
			c.Logger.Panic(err)
		}

		if rCheckGuildName.GetGuildNameFree() || c.Update.Message.Text == "" {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.guild_name_not_exists")
			return true
		}

		c.Payload.Name = c.Update.Message.Text
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
func (c *SafePlanetGuildJoinController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player a quale gilda vuole unirsi
	// ##################################################################################################
	case 0:
		// Recupero lista di gilde
		var rGetJoinGuildsList *pb.GetJoinGuildsListResponse
		if rGetJoinGuildsList, err = config.App.Server.Connection.GetJoinGuildsList(helpers.NewContext(1), &pb.GetJoinGuildsListRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		var guildKeyboard [][]tgbotapi.KeyboardButton
		for _, guild := range rGetJoinGuildsList.GetGuildsList() {
			guildKeyboard = append(guildKeyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					guild.GetName(),
				),
			))
		}

		// Aggiungo torna al menu
		guildKeyboard = append(guildKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		))

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.join_start"))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       guildKeyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1
	// ##################################################################################################
	// Chiedo Conferma al player se vuole entrare nella gilda indicata
	// ##################################################################################################
	case 1:
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.join_confirm", c.Payload.Name))
		msg.ParseMode = "markdown"
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
		_, err = config.App.Server.Connection.JoinGuild(helpers.NewContext(1), &pb.JoinGuildRequest{
			PlayerID:  c.Player.ID,
			GuildName: c.Payload.Name,
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

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.guild.join_completed_ok"))
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
