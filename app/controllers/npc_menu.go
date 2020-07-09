package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Writer: reloonfire
// Starting on: 09/07/2020
// Project: nn-telegram

type NpcMenuController struct {
	BaseController
	SafePlanet bool // Flag per verificare se il player si trova su un pianeta sicuro
}

// ====================================
// Handle
// ====================================
func (c *NpcMenuController) Handle(player nnsdk.Player, update tgbotapi.Update, proxy bool) {
	var err error
	var playerProvider providers.PlayerProvider

	// Il menù del player refresha sempre lo status del player
	player, err = playerProvider.FindPlayerByUsername(player.Username)
	if err != nil {
		panic(err)
	}

	// Init funzionalità
	c.Controller = "route.menu.npc"
	c.Player = player

	msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.welcome"))
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       c.GetKeyboard(),
		ResizeKeyboard: true,
	}

	// Send recap message
	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}

	// Se il player è morto non può fare altro che riposare o azioni che richiedono riposo
	if *c.Player.Stats.Dead {
		restsController := new(ShipRestsController)
		restsController.Handle(c.Player, c.Update, true)
	}
}

func (c *NpcMenuController) GetKeyboard() [][]tgbotapi.KeyboardButton {
	var npcProvider providers.NpcProvider

	// Recupero gli npc attivi in questo momento
	npcs, err := npcProvider.GetAll()
	if err != nil {
		panic(err)
	}

	var keyboardRow [][]tgbotapi.KeyboardButton
	for _, npc := range npcs {
		row := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("route.safeplanet.%s", npc.Slug)),
			),
		)
		keyboardRow = append(keyboardRow, row)
	}

	keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.menu")),
	))

	return keyboardRow
}
