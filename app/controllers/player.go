package controllers

import (
	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Player
// ====================================
type PlayerController struct {
	Payload interface{}
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *PlayerController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error
	c.Player = player
	c.Update = update
	c.Configuration.Controller = "route.player"

	// *************************
	// Recupero economia player
	// *************************
	var money, diamond int32
	money, diamond, err = c.GetPlayerEconomy()
	if err != nil {
		panic(err)
	}

	recapPlayer := helpers.Trans(
		c.Player.Language.Slug,
		"player.datails.card",
		c.Player.GetUsername(),
		c.PlayerData.PlayerStats.GetLifePoint(),
		c.PlayerData.PlayerStats.GetExperience(),
		c.PlayerData.PlayerStats.GetLevel(),
		money, diamond,
	)

	// *************************
	// Recupero quanti pianeti ha visitato
	// *************************
	var rCountPlanetVisited *pb.CountPlanetVisitedResponse
	rCountPlanetVisited, err = services.NnSDK.CountPlanetVisited(helpers.NewContext(100), &pb.CountPlanetVisitedRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		// log.Fatalln(err)
		panic(err)
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
	rCountSystemVisited, err = services.NnSDK.CountSystemVisited(helpers.NewContext(100), &pb.CountSystemVisitedRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		// log.Fatalln(err)
		panic(err)
	}

	recapPlayer += helpers.Trans(
		c.Player.Language.Slug,
		"player.datails.system_visited",
		rCountSystemVisited.GetValue(),
	)

	// msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "player.intro"))
	msg := services.NewMessage(c.Update.Message.Chat.ID, recapPlayer)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory.equip")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
		),
	)

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}

func (c *PlayerController) Validator() {
	//
}

func (c *PlayerController) Stage() {
	//
}

// GetPlayerTask
// Metodo didicato alla reppresenteazione del risorse econimiche del player
func (c *PlayerController) GetPlayerEconomy() (money int32, diamond int32, err error) {
	// Calcolo monete del player
	responseMoney, _ := services.NnSDK.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
		PlayerID:    c.Player.GetID(),
		EconomyType: "money",
	})

	// Calcolo diamanti del player
	responseDiamond, _ := services.NnSDK.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
		PlayerID:    c.Player.GetID(),
		EconomyType: "diamond",
	})

	return responseMoney.GetValue(), responseDiamond.GetValue(), nil
}
