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
	var playerProvider providers.PlayerProvider

	// Il menù del player refresha sempre lo status del player
	player, err = playerProvider.FindPlayerByUsername(player.Username)
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

	// Se il player è morto non può fare altro che riposare o azioni che richiedono riposo
	if *c.Player.Stats.Dead {
		restsController := new(ShipRestsController)
		restsController.Handle(c.Player, c.Update)
	}
}

// GetRecap
// BoardSystem v0.1
// 🌏 Nomepianeta
// 👨🏼‍🚀 Casteponters
// ♥️ life/max-life
// 💰 XXXX 💎 XXXX
//
// ⏱ Task in corso:
// - LIST
func (c *MenuController) GetRecap() (message string, err error) {
	var playerProvider providers.PlayerProvider
	var planetProvider providers.PlanetProvider

	// Recupero ultima posizione del player, dando per scontato che sia
	// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
	var lastPosition nnsdk.PlayerPosition
	lastPosition, err = playerProvider.GetPlayerLastPosition(c.Player)
	if err != nil {
		return message, err
	}

	// Dalla ultima posizione recupero il pianeta corrente
	var planet nnsdk.Planet
	planet, err = planetProvider.GetPlanetByCoordinate(lastPosition.X, lastPosition.Y, lastPosition.Z)
	if err != nil {
		return message, err
	}

	// Calcolo lato economico del player
	var economy string
	economy, err = c.GetPlayerEconomy()
	if err != nil {
		return message, err
	}

	// Recupero status vitale del player
	var life string
	life, err = c.GetPlayerLife()
	if err != nil {
		return message, err
	}

	message = helpers.Trans(c.Player.Language.Slug, "menu",
		planet.Name,
		c.Player.Username,
		life,
		economy,
		c.GetPlayerTasks(),
	)

	return
}

// GetPlayerTask
func (c *MenuController) GetPlayerEconomy() (economy string, err error) {
	var playerProvider providers.PlayerProvider

	// Calcolo monete del player
	var money nnsdk.MoneyResponse
	money, _ = playerProvider.GetPlayerEconomy(c.Player.ID, "money")

	var diamond nnsdk.MoneyResponse
	diamond, _ = playerProvider.GetPlayerEconomy(c.Player.ID, "diamond")

	economy = fmt.Sprintf("💰 %v 💎 %v", money.Value, diamond.Value)

	return
}

// GetPlayerLife
func (c *MenuController) GetPlayerLife() (life string, err error) {
	// Calcolo stato vitale del player
	status := "♥️"
	if *c.Player.Stats.Dead {
		status = "☠️"
	}

	life = fmt.Sprintf("%s️ %v/%v HP", status, *c.Player.Stats.LifePoint, 100)

	return
}

// GetPlayerTask
func (c *MenuController) GetPlayerTasks() (tasks string) {
	if len(c.Player.States) > 0 {
		tasks = helpers.Trans(c.Player.Language.Slug, "menu.tasks")

		for _, state := range c.Player.States {
			if !*state.Completed {
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
	}

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
		} else if state.Controller == "route.mission" {
			return c.MissionKeyboard()
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
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ability")),
		},
	}
}

// TutorialMenu
func (c *MenuController) TutorialKeyboard() (keyboardRows [][]tgbotapi.KeyboardButton) {
	// Per il tutorial costruisco keyboard solo per gli stati attivi
	for _, state := range c.Player.States {
		if !*state.Completed {
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
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ability")),
		},
	}
}

// MissionKeyboard
func (c *MenuController) MissionKeyboard() [][]tgbotapi.KeyboardButton {
	return [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.mission")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.inventory")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.crafting")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ability")),
		},
	}
}
