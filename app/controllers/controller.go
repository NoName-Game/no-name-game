package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	// Lista delle rotte che non devono subire
	// gli effetti di abbandona anche se forzati a mano
	UnClearables = []string{"route.hunting"}
)

type BaseController struct {
	Update     tgbotapi.Update
	Controller string
	Father     uint
	Validation struct {
		HasErrors     bool
		Message       string
		ReplyKeyboard tgbotapi.ReplyKeyboardMarkup
	}
	Player nnsdk.Player
	State  nnsdk.PlayerState
}

// Completing - Metodo per settare il completamento di uno stato
func (c *BaseController) Completing() (err error) {
	var playerStateProvider providers.PlayerStateProvider

	// Verifico se lo stato è completato chiudo
	if *c.State.Completed == true {
		// Posso cancellare lo stato solo se non è figlio di qualche altro stato
		if c.State.Father <= 0 {
			_, err = playerStateProvider.DeletePlayerState(c.State) // Delete
			if err != nil {
				return err
			}
		}

		// Cancello stato da redis
		err = helpers.DelRedisState(c.Player)
		if err != nil {
			panic(err)
		}

		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update)
	}

	return
}

// Clearable
func (c *BaseController) Clearable() (clearable bool) {
	// Certi controller non devono subire la cancellazione degli stati
	// perchè magari hanno logiche particolari o lo gestiscono a loro modo
	for _, state := range c.Player.States {
		for _, unclearable := range UnClearables {
			if helpers.Trans(c.Player.Language.Slug, state.Controller) == helpers.Trans(c.Player.Language.Slug, unclearable) {
				return false
			}
		}
	}

	return true
}
