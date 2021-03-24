package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerGuildController
// ====================================
type PlayerGuildController struct {
	Controller
}

func (c *PlayerGuildController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.guild",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerController{},
				FromStage: 0,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *PlayerGuildController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Recupero gilda player
	var rGetPlayerGuild *pb.GetPlayerGuildResponse
	if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	// #####################################
	// Se il player non appartiene a nessuna gilda
	// #####################################
	if !rGetPlayerGuild.GetInGuild() {
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "player.guild.player_not_in_one_guild"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		return
	}

	// #####################################
	// Recupero owner gilda
	// #####################################
	var rGetPlayerByID *pb.GetPlayerByIDResponse
	if rGetPlayerByID, err = config.App.Server.Connection.GetPlayerByID(helpers.NewContext(1), &pb.GetPlayerByIDRequest{
		ID: rGetPlayerGuild.GetGuild().GetOwnerID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// #####################################
	// Recupero punti gilda
	// #####################################
	var rGetGuildPoints *pb.GetGuildPointsResponse
	if rGetGuildPoints, err = config.App.Server.Connection.GetGuildPoints(helpers.NewContext(1), &pb.GetGuildPointsRequest{
		GuildID: rGetPlayerGuild.GetGuild().GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Recap messagi
	guildDetails := helpers.Trans(player.Language.Slug, "player.guild.guild_details",
		rGetPlayerGuild.GetGuild().GetName(),     // Nome
		rGetGuildPoints.GetResult(),              // Punti
		rGetPlayerByID.GetPlayer().GetUsername(), // Fondatore
	)

	// #####################################
	// Recupero lista partecipanti gilda
	// #####################################
	var rGetPlayersGuild *pb.GetPlayersGuildResponse
	if rGetPlayersGuild, err = config.App.Server.Connection.GetPlayersGuild(helpers.NewContext(1), &pb.GetPlayersGuildRequest{
		GuildID: rGetPlayerGuild.GetGuild().GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	var playersList string
	playersList = helpers.Trans(player.Language.Slug, "player.guild.members")
	for _, playerDetails := range rGetPlayersGuild.GetPlayers() {
		// Recupero posizione player corrente
		var playerPosition *pb.Planet
		if playerPosition, err = helpers.GetPlayerPosition(playerDetails.GetID()); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero punti player
		var rGetPlayerGuildPoints *pb.GetPlayerGuildPointsResponse
		if rGetPlayerGuildPoints, err = config.App.Server.Connection.GetPlayerGuildPoints(helpers.NewContext(1), &pb.GetPlayerGuildPointsRequest{
			GuildID:  rGetPlayerGuild.GetGuild().GetID(),
			PlayerID: playerDetails.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		var rGetRankByID *pb.GetRankByIDResponse
		if rGetRankByID, err = config.App.Server.Connection.GetRankByID(helpers.NewContext(1), &pb.GetRankByIDRequest{
			RankID: playerDetails.GetRankID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		playersList += helpers.Trans(player.Language.Slug, "player.guild.player_details",
			playerDetails.GetUsername(),
			playerPosition.GetName(),
			playerDetails.GetLevelID(),
			helpers.Trans(c.Player.GetLanguage().GetSlug(), fmt.Sprintf("rank.%s", rGetRankByID.GetRank().GetNameCode())),
			rGetPlayerGuildPoints.GetResult(),
		)
	}

	// #####################################
	// Invio messggio recap
	// #####################################
	msg := helpers.NewMessage(c.ChatID, fmt.Sprintf("%s\n\n%s",
		guildDetails,
		playersList,
	))
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *PlayerGuildController) Validator() bool {
	return false
}

func (c *PlayerGuildController) Stage() {
	//
}
