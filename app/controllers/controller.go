package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

func (c *BaseController) InitController(controller string, player nnsdk.Player) (err error) {
	// TODO: portare qui i controlli iniziali

	// Se il player è morto lo mando a riposare
	if *player.Stats.Dead && controller != "route.ship.rests" {
		restsController := new(ShipRestsController)
		restsController.Handle(c.Player, c.Update)
	}

	return
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
