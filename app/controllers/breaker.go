package controllers

import (
	"os"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BreakerController struct {
	BaseController
}

// BackController
// Back permette la cancellazione dello stato solo da REDIS, e NON quello a DB
// questo permetterà di tornare al menù ed eseugire altre operazioni, così che quando
// l'utente ritornerà sulla funzionalità precendete potrà riprendere da dove aveva lasciato
type BackController BreakerController

// ====================================
// Handle
// ====================================
func (c *BackController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	var err error

	// Inizializzo
	c.Controller = "route.breaker.back"
	c.Message = update.Message

	// Delete redis state
	err = helpers.DelRedisState(player)
	if err != nil {
		panic(err)
	}

	// Questo messaggio verrà mostrato solo in vase di debug
	if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
		msg := services.NewMessage(player.ChatID,
			"***************************\nDEBUG: DELETE REDIS STATE.\n***************************\n",
		)
		// msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, err = services.SendMessage(msg)
		if err != nil {
			panic(err)
		}
	}

	// Call menu controller
	new(MenuController).Handle(player, update)
}

// ClearsController
// Come il backcontroller solo che questa funzionalità permette
// anche la cancellazione in soft-delete del record dal DB, questo permette
// di poter ricominciare da capo una funzionalità
type ClearsController BreakerController

// ====================================
// Handle
// ====================================
func (c *ClearsController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	var err error

	// Inizializzo
	c.Controller = "route.breaker.back"
	c.Message = update.Message

	err = helpers.DeleteRedisAndDbState(player)
	if err != nil {
		panic(err)
	}

	message := update.Message
	if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
		msg := services.NewMessage(message.Chat.ID,
			"***************************\nDEBUG: DELETE DB AND REDIS STATE.\n***************************\n",
		)
		// msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, err = services.SendMessage(msg)
		if err != nil {
			panic(err)
		}
	}

	// Call menu controller
	new(MenuController).Handle(player, update)
}
