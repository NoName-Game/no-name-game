package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type MenuController BaseController

// ====================================
// Handle
// ====================================
func (c *MenuController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	var err error

	// Il menù del player refresha sempre lo status del player
	player, err = providers.FindPlayerByUsername(player.Username)
	if err != nil {
		panic(err)
	}

	// Init funzionalità
	c.Controller = "route.menu"
	c.Player = player

	// Recupero messaggio principale
	var recap string
	recap, err = c.GetRecap()
	if err != nil {
		panic(err)
	}

	msg := services.NewMessage(c.Player.ChatID, recap)
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
}

// GetPlayerTask
func (c *MenuController) GetPlayerTasks() (tasks string) {
	for _, state := range c.Player.States {
		if *state.Completed != true {
			// Se sono da notificare formatto con la data
			if *state.ToNotify {
				tasks += fmt.Sprintf("- %s (%s)\n",
					helpers.Trans(c.Player.Language.Slug, state.Controller),
					state.FinishAt.Format("15:04:05 01/02"),
				)
			} else {
				tasks += fmt.Sprintf("- %s\n", helpers.Trans(c.Player.Language.Slug, state.Controller))
			}
		}
	}

	return
}

// GetRecap
// BoardSystem v0.1
// 🌏 Nomepianeta
// 👨🏼‍🚀 Casteponters
// ♥️ life/max-life
//
// ⏱ Task in corso:
// - LIST
func (c *MenuController) GetRecap() (message string, err error) {

	// Recupero ultima posizione del player, dando per scontato che sia
	// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
	var lastPosition nnsdk.PlayerPosition
	lastPosition, err = providers.GetPlayerLastPosition(c.Player)
	if err != nil {
		return message, err
	}

	// Dalla ultima posizione recupero il pianeta corrente
	var planet nnsdk.Planet
	planet, err = providers.GetPlanetByCoordinate(lastPosition.X, lastPosition.Y, lastPosition.Z)
	if err != nil {
		return message, err
	}

	message = helpers.Trans(c.Player.Language.Slug, "menu",
		planet.Name,
		c.Player.Username,
		*c.Player.Stats.LifePoint, 100,
		c.GetPlayerTasks(),
	)

	return
}

// GetRecap
func (c *MenuController) GetKeyboard() [][]tgbotapi.KeyboardButton {
	// Se il player sta finendo il tutorial mostro il menù con i task personalizzati
	// var inTutorial bool
	for _, state := range c.Player.States {
		if state.Controller == "route.tutorial" {
			return c.TutorialKeyboard()
		} else if state.Controller == "route.ship.exploration" {
			return c.ExplorationKeyboard()
		}
	}

	return c.MainKeyboard()
}

// MainMenu
func (c *MenuController) MainKeyboard() [][]tgbotapi.KeyboardButton {
	return [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.mission")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.hunting")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.inventory")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.crafting")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.abilityTree")),
		},
	}
}

// TutorialMenu
func (c *MenuController) TutorialKeyboard() (keyboardRows [][]tgbotapi.KeyboardButton) {
	// Per il tutorial costruisco keyboard solo per gli stati attivi
	for _, state := range c.Player.States {
		if *state.Completed != true {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, state.Controller)),
			)

			keyboardRows = append(keyboardRows, keyboardRow)
		}
	}

	return
}

// ExplorationKeyboard
func (c *MenuController) ExplorationKeyboard() [][]tgbotapi.KeyboardButton {
	return [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.inventory")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.crafting")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.abilityTree")),
		},
	}
}
