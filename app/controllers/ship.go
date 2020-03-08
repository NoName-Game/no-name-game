package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
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
func (c *ShipController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	var playerProvider providers.PlayerProvider

	c.Controller = "route.ship"
	c.Player = player
	c.Update = update

	// Se tutto ok imposto e setto il nuovo stato su redis
	_ = helpers.SetRedisState(c.Player, c.Controller)

	if c.Clear() {
		return
	}

	// Recupero nave attiva de player
	var eqippedShips nnsdk.Ships
	eqippedShips, err = playerProvider.GetPlayerShips(c.Player, true)
	if err != nil {
		panic(err)
	}

	var currentShipRecap string
	for _, ship := range eqippedShips {
		currentShipRecap = fmt.Sprintf(
			"🚀 %s (%s)\n🏷 %s\n🔧 %v%% (%s)\n⛽ %v%% (%s)",
			ship.Name, ship.Rarity.Slug,
			ship.ShipCategory.Name,
			ship.ShipStats.Integrity, helpers.Trans(c.Player.Language.Slug, "integrity"),
			*ship.ShipStats.Tank, helpers.Trans(c.Player.Language.Slug, "fuel"),
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
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		),
	)

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}
