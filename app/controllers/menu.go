package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Example:
// Computer di bordo di reloonfire
// Task in corso:
// Tutorial.
// Crafting: Termina alle ore xx:xx:xx.
// Versione di sviluppo di NoNameGame, tutti i testi potranno cambiare con la release ufficiale.

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

	// Inizializzo
	c.Controller = "route.menu"
	c.Player = player

	// Keyboard menu
	var keyboardMenu = [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.mission")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.hunting")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.inventory")),
		},
		{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.crafting")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.abilityTree")),
		},
	}

	var tasks string
	var keyboardRows [][]tgbotapi.KeyboardButton

	for _, state := range c.Player.States {
		if *state.Completed != true {
			if *state.ToNotify {
				// If FinishAt is setted "On Going %TASKNAME: Finish at XX:XX:XX"
				stateText := helpers.Trans(c.Player.Language.Slug, state.Controller) + helpers.Trans(c.Player.Language.Slug, "menu.finishAt", state.FinishAt.Format("15:04:05"))
				tasks += helpers.Trans(c.Player.Language.Slug, "menu.onGoing", stateText) + "\n"
			} else {
				tasks += helpers.Trans(c.Player.Language.Slug, "menu.onGoing", helpers.Trans(c.Player.Language.Slug, state.Controller)) + "\n"
			}

			keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, state.Controller)))
			keyboardRows = append(keyboardRows, keyboardRow)
		}
	}

	msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "menu", c.Player.Username, tasks))
	msg.ParseMode = "HTML"

	var inTutorial bool
	for _, state := range c.Player.States {
		// Se il player sta finendo il tutorial mostro il menù con i task personalizzati
		if state.Controller == "route.tutorial" {
			inTutorial = true
			break
		}
	}

	// Verifico se è in tutorial o no
	if inTutorial {
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRows,
			ResizeKeyboard: true,
		}
	} else {
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardMenu,
			ResizeKeyboard: true,
		}
	}

	// Send recap message
	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}
