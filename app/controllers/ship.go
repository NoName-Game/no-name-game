package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipController
// Ogni player ha la possibilità di spostarsi nei diversi pianeti
// del sistema di NoName
// ====================================
type ShipController struct {
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *ShipController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	c.Controller = "route.ship"
	c.Player = player
	c.Update = update

	// Se tutto ok imposto e setto il nuovo stato su redis
	_ = helpers.SetRedisState(*c.Player, c.Controller)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &MenuController{}) {
			return
		}
	}

	// Recupero nave attiva de player
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := services.NnSDK.GetPlayerShips(ctx, &pb.GetPlayerShipsRequest{
		PlayerID: c.Player.GetID(),
		Equipped: true,
	})
	if err != nil {
		panic(err)
	}
	var eqippedShips []*pb.Ship
	eqippedShips = response.GetShips()

	var currentShipRecap string
	for _, ship := range eqippedShips {
		currentShipRecap = fmt.Sprintf(
			"🚀 %s (%s)\n🏷 %s\n🔧 %v%% (%s)\n⛽ %v%% (%s)",
			ship.Name, strings.ToUpper(ship.Rarity.Slug),
			ship.ShipCategory.Name,
			ship.ShipStats.Integrity, helpers.Trans(c.Player.Language.Slug, "integrity"),
			ship.ShipStats.Tank, helpers.Trans(c.Player.Language.Slug, "fuel"),
		)
	}

	// Invio messaggio
	msg := services.NewMessage(c.Update.Message.Chat.ID,
		fmt.Sprintf(
			"%s:\n\n %s",
			helpers.Trans(c.Player.Language.Slug, "ship.report"),
			currentShipRecap,
		),
	)

	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.exploration")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.rests")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.repairs")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.crafting")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		),
	)

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}

func (c *ShipController) Validator() {
	//
}

func (c *ShipController) Stage() {
	//
}
