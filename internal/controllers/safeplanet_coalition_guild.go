package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetProtectorsController
// ====================================
type SafePlanetProtectorsController struct {
	Controller
}

func (c *SafePlanetProtectorsController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.protectors",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
			AllowedControllers: []string{
				"route.safeplanet.coalition.protectors.create",
				"route.safeplanet.coalition.protectors.join",
				"route.safeplanet.coalition.protectors.leave",
				"route.safeplanet.coalition.protectors.add_player",
				"route.safeplanet.coalition.protectors.remove_player",
				"route.safeplanet.coalition.protectors.change_leader",
				"route.safeplanet.coalition.protectors.switch",
				"route.safeplanet.coalition.protectors.change_name",
				"route.safeplanet.coalition.protectors.change_tag",
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetProtectorsController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Verifico/Recupero Gilda player
	var err error
	var rGetPlayerGuild *pb.GetPlayerGuildResponse
	if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	var keyboardRows [][]tgbotapi.KeyboardButton
	if !rGetPlayerGuild.GetInGuild() {
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.create")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.join")),
		))
	} else {
		// Se il player è owner dell gruppo allora vede il tasto per gestire gli altri player
		if rGetPlayerGuild.GetGuild().GetOwnerID() == c.Player.ID {
			keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
				// tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.add_player")),
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.switch")),
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.remove_player")),
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.change_leader")),
			))

			keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.change_name")),
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.change_tag")),
			))
		}

		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.leave")),
		))
	}

	// Aggiungo anche abbandona
	keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(
			helpers.Trans(c.Player.Language.Slug, "route.breaker.menu"),
		),
	))

	// Recupero lista partecipanti gilda
	var rGetPlayersGuild *pb.GetPlayersGuildResponse
	if rGetPlayersGuild, err = config.App.Server.Connection.GetPlayersGuild(helpers.NewContext(1), &pb.GetPlayersGuildRequest{
		GuildID: rGetPlayerGuild.GetGuild().GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Recupero punti gilda
	var rGetGuildPoints *pb.GetGuildPointsResponse
	if rGetGuildPoints, err = config.App.Server.Connection.GetGuildPoints(helpers.NewContext(1), &pb.GetGuildPointsRequest{
		GuildID: rGetPlayerGuild.GetGuild().GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Message
	var protectorsMessage string
	protectorsMessage = helpers.Trans(player.Language.Slug, "safeplanet.coalition.protectors.info")
	if rGetPlayerGuild.GetInGuild() {
		protectorsMessage += helpers.Trans(player.Language.Slug, "safeplanet.coalition.protectors.player_protectors_details",
			rGetPlayerGuild.GetGuild().GetName(),
			rGetPlayerGuild.GetGuild().GetTag(),
			len(rGetPlayersGuild.GetPlayers()),
			rGetGuildPoints.GetResult(),
		)
	}

	msg := helpers.NewMessage(c.ChatID, protectorsMessage)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       keyboardRows,
		ResizeKeyboard: true,
	}

	if _, err := helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *SafePlanetProtectorsController) Validator() bool {
	return false
}

func (c *SafePlanetProtectorsController) Stage() {
	//
}
