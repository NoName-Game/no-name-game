package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BaseController struct {
	Update     tgbotapi.Update
	Controller string
	Father     uint
	Validation struct {
		HasErrors bool
		Message   string
	}
	Player nnsdk.Player
	State  nnsdk.PlayerState
}

// Completing - Metodo per settare il completamento di uno stato
func (c *BaseController) Completing() (err error) {
	// Verifico se lo stato è completato chiudo
	if *c.State.Completed == true {
		// Posso cancellare lo stato solo se non è figlio di qualche altro stato
		if c.State.Father <= 0 {
			_, err = providers.DeletePlayerState(c.State) // Delete
			if err != nil {
				return err
			}
		}

		// Cancello stato da redis
		err = helpers.DelRedisState(c.Player)
		if err != nil {
			panic(err)
		}

		// Se si trova in hunting aggiunto un controllo
		if c.Controller == "route.hunting" {
			// Cancello messaggio contentente la mappa accertandomi che l'azione
			// arrivi da un messaggio di callback
			if c.Update.CallbackQuery != nil {
				err = services.DeleteMessage(c.Update.CallbackQuery.Message.Chat.ID, c.Update.CallbackQuery.Message.MessageID)
				if err != nil {
					return err
				}
			}
		}

		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update)
	}

	return
}
