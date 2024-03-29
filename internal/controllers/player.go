package controllers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"
)

// ====================================
// Player
// ====================================
type PlayerController struct {
	Controller
}

func (c *PlayerController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
			AllowedControllers: []string{
				"route.player.inventory",
				"route.player.inventory.equip",
				"route.player.guild",
				"route.player.party",
				"route.player.achievements",
				"route.player.emblems",
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *PlayerController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// *************************
	// Recupero economia player
	// *************************
	var money, diamond int32
	money, diamond = c.GetPlayerEconomy()

	// *************************
	// Recupero esperienza player
	// *************************
	var rGetPlayerExperience *pb.GetPlayerExperienceResponse
	if rGetPlayerExperience, err = config.App.Server.Connection.GetPlayerExperience(helpers.NewContext(1), &pb.GetPlayerExperienceRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Recupero dettagli livello sucessivo
	var rGetLevelByID *pb.GetLevelByIDResponse
	rGetLevelByID, _ = config.App.Server.Connection.GetLevelByID(helpers.NewContext(1), &pb.GetLevelByIDRequest{
		LevelID: c.Player.GetLevelID() + 1,
	})

	// *************************
	// Recupero amuleti player
	// *************************
	var rGetPlayerItems *pb.GetPlayerItemsResponse
	if rGetPlayerItems, err = config.App.Server.Connection.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// *************************
	// Recupero rank player
	// *************************
	var rGetPlayerRankPoint *pb.GetPlayerRankPointResponse
	if rGetPlayerRankPoint, err = config.App.Server.Connection.GetPlayerRankPoint(helpers.NewContext(1), &pb.GetPlayerRankPointRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	var rGetRankByID *pb.GetRankByIDResponse
	rGetRankByID, _ = config.App.Server.Connection.GetRankByID(helpers.NewContext(1), &pb.GetRankByIDRequest{RankID: c.Player.RankID + 1})

	var amulets int32
	for _, item := range rGetPlayerItems.GetPlayerInventory() {
		if item.Item.ID == 7 {
			amulets = item.Quantity
		}
	}

	rank := helpers.Trans(c.Player.Language.Slug, "rank."+c.Player.Rank.NameCode)

	// *************************
	// Recupero gilda player
	// *************************
	var rGetPlayerGuild *pb.GetPlayerGuildResponse
	if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	var guildName = "--"
	if rGetPlayerGuild.GetGuild().GetID() > 0 {
		guildName = rGetPlayerGuild.GetGuild().GetName()
	}

	recapPlayer := helpers.Trans(
		c.Player.Language.Slug,
		"player.datails.card",
		c.Player.GetUsername(),
		guildName,
		c.Player.GetLifePoint(),                                                         // Life point player
		c.Player.GetLevel().GetPlayerMaxLife(),                                          // Vita massima del player
		rGetPlayerExperience.GetValue(), rGetLevelByID.GetLevel().GetExperienceNeeded(), // Esperienza
		c.Player.GetLevel().GetID(), // Livello
		rank,
		rGetPlayerRankPoint.GetValue(), rGetRankByID.GetRank().GetPointNeeded(),
		money, diamond, amulets,
	)

	// *************************
	// Recupero quanti pianeti ha visitato
	// *************************
	var rCountPlanetVisited *pb.CountPlanetVisitedResponse
	if rCountPlanetVisited, err = config.App.Server.Connection.CountPlanetVisited(helpers.NewContext(100), &pb.CountPlanetVisitedRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	recapPlayer += helpers.Trans(
		c.Player.Language.Slug,
		"player.datails.planet_visited",
		rCountPlanetVisited.GetValue(),
	)

	// *************************
	// Recupero quanti sistemi ha visitatao
	// *************************
	var rCountSystemVisited *pb.CountSystemVisitedResponse
	if rCountSystemVisited, err = config.App.Server.Connection.CountSystemVisited(helpers.NewContext(100), &pb.CountSystemVisitedRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	recapPlayer += helpers.Trans(
		c.Player.Language.Slug,
		"player.datails.system_visited",
		rCountSystemVisited.GetValue(),
	)

	// msg := services.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "player.intro"))
	msg := helpers.NewMessage(c.ChatID, recapPlayer)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.inventory")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.inventory.equip")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.guild")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.party")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.achievements")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.player.emblems")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *PlayerController) Validator() bool {
	return false
}

func (c *PlayerController) Stage() {
	//
}

// GetPlayerTask
// Metodo didicato alla reppresenteazione del risorse econimiche del player
func (c *PlayerController) GetPlayerEconomy() (money int32, diamond int32) {
	var err error

	// Calcolo monete del player
	var rGetPlayerEconomyMoney *pb.GetPlayerEconomyResponse
	if rGetPlayerEconomyMoney, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
		PlayerID:    c.Player.GetID(),
		EconomyType: pb.GetPlayerEconomyRequest_MONEY,
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Calcolo diamanti del player
	var rGetPlayerEconomyDiamond *pb.GetPlayerEconomyResponse
	if rGetPlayerEconomyDiamond, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
		PlayerID:    c.Player.GetID(),
		EconomyType: pb.GetPlayerEconomyRequest_DIAMOND,
	}); err != nil {
		c.Logger.Panic(err)
	}

	return rGetPlayerEconomyMoney.GetValue(), rGetPlayerEconomyDiamond.GetValue()
}
