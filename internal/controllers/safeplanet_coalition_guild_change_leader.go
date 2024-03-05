package controllers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"
)

// ====================================
// SafePlanetProtectorsChangeLeaderController
// ====================================
type SafePlanetProtectorsChangeLeaderController struct {
	Payload struct {
		Username string
	}
	Controller
}

func (c *SafePlanetProtectorsChangeLeaderController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.protectors.change_leader",
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
func (c *SafePlanetProtectorsChangeLeaderController) Handle(player *pb.Player, update tgbotapi.Update) {
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
func (c *SafePlanetProtectorsChangeLeaderController) Validator() bool {
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
	// Verifico se il player scelto esisteo
	// ##################################################################################################
	case 1:
		// Recupero gilda player
		var rGetPlayerGuild *pb.GetPlayerGuildResponse
		if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Ciclo tutta la lista degli esploratori nella gilda
		var rGetPlayersGuild *pb.GetPlayersGuildResponse
		if rGetPlayersGuild, err = config.App.Server.Connection.GetPlayersGuild(helpers.NewContext(1), &pb.GetPlayersGuildRequest{
			GuildID: rGetPlayerGuild.GetGuild().GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		var found bool
		for _, player := range rGetPlayersGuild.GetPlayers() {
			if player.GetUsername() == c.Update.Message.Text {
				found = true
			}
		}

		if !found || c.Update.Message.Text == "" {
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
func (c *SafePlanetProtectorsChangeLeaderController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player di indicare quale player vuole aggiungere alla sua gilda
	// ##################################################################################################
	case 0:
		// Verifico/Recupero Gilda player
		var err error
		var rGetPlayerGuild *pb.GetPlayerGuildResponse
		if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		var protectorsKeyboard [][]tgbotapi.KeyboardButton

		// Ciclo tutta la lista degli esploratori nella gilda
		var rGetPlayersGuild *pb.GetPlayersGuildResponse
		if rGetPlayersGuild, err = config.App.Server.Connection.GetPlayersGuild(helpers.NewContext(1), &pb.GetPlayersGuildRequest{
			GuildID: rGetPlayerGuild.GetGuild().GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		for _, player := range rGetPlayersGuild.GetPlayers() {
			if player.GetID() != c.Player.GetID() {
				protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(player.GetUsername()),
				))
			}
		}

		// Aggiungo torna al menu
		protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.change_leader_start"))
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
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.change_leader_confirm", c.Payload.Username))
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

		if _, err = config.App.Server.Connection.ChangeLeaderGuild(helpers.NewContext(1), &pb.ChangeLeaderGuildRequest{
			PlayerUsername: c.Payload.Username,
			GuildID:        rGetPlayerGuild.GetGuild().GetID(),
		}); err != nil {
			c.Logger.Warning(err)

			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.change_leader_completed_ko"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		}

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.change_leader_completed_ok"))
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
