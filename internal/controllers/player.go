package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Player
// ====================================
type PlayerController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(Controller{
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
		},
	}) {
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

	// *************************
	// Recupero amuleti player
	// *************************
	var rGetPlayerItems *pb.GetPlayerItemsResponse
	if rGetPlayerItems, err = config.App.Server.Connection.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	var amulets int32
	for _, item := range rGetPlayerItems.GetPlayerInventory() {
		if item.Item.ID == 7 {
			amulets = item.Quantity
		}
	}

	recapPlayer := helpers.Trans(
		c.Player.Language.Slug,
		"player.datails.card",
		c.Player.GetUsername(),
		c.Player.GetLifePoint(),                // Life point player
		c.Player.GetLevel().GetPlayerMaxLife(), // Vita massima del player
		rGetPlayerExperience.GetValue(),        // Esperienza
		c.Player.GetLevel().GetID(),            // Livello
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

	// msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "player.intro"))
	msg := helpers.NewMessage(c.Update.Message.Chat.ID, recapPlayer)
	msg.ParseMode = tgbotapi.ModeMarkdown
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
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
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
