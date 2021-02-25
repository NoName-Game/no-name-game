package controllers

import (
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetProtectorsJoinController
// ====================================
type SafePlanetProtectorsJoinController struct {
	Payload struct {
		Name string
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetProtectorsJoinController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.protectors.join",
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
func (c *SafePlanetProtectorsJoinController) Validator() bool {
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
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.protectors_name_not_exists")
			return true
		}

		var rGetGuildByName *pb.GetGuildByNameResponse
		if rGetGuildByName, err = config.App.Server.Connection.GetGuildByName(helpers.NewContext(1), &pb.GetGuildByNameRequest{GuildName: c.Update.Message.Text}); err != nil {
			c.Logger.Panic(err)
		}
		if rGetGuildByName.GetGuild().GetGuildType() {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.cannot_join")
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu"))))
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
func (c *SafePlanetProtectorsJoinController) Stage() {
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

		var protectorsKeyboard [][]tgbotapi.KeyboardButton
		for _, protectors := range rGetJoinGuildsList.GetGuildsList() {
			protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					protectors.GetName(),
				),
			))
		}

		// Aggiungo torna al menu
		protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.join_start"))
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
	// Chiedo Conferma al player se vuole entrare nella gilda indicata
	// ##################################################################################################
	case 1:
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.join_confirm", c.Payload.Name))
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
		_, err = config.App.Server.Connection.JoinGuild(helpers.NewContext(1), &pb.JoinGuildRequest{
			PlayerID:  c.Player.ID,
			GuildName: c.Payload.Name,
		})

		if err != nil && strings.Contains(err.Error(), "player already in one guild") {
			errorMsg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.player_already_in_one_protectors"))
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		} else if err != nil && strings.Contains(err.Error(), "error guild limit reached") {
			errorMsg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.max_player_reached"))
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		}

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.join_completed_ok"))
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
