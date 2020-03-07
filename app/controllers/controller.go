package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/services"

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
	ToMenu bool
}

// Completing - Metodo per settare il completamento di uno stato
func (c *BaseController) Completing() (err error) {
	var playerStateProvider providers.PlayerStateProvider

	// Verifico se lo stato è completato chiudo
	if *c.State.Completed {
		// Posso cancellare lo stato solo se non è figlio di qualche altro stato
		if c.State.Father == 0 {
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

		return
	}

	// Verifico se si vuole forzare il menu
	if c.ToMenu {
		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update)
	}

	return
}

// InStatesBlocker
// Certi controller non possono essere eseguiti se il player si trova in determinati stati.
// Ogni controller ha la possibilità nell'handle di passare la lista di rotte bloccanti per esso.
func (c *BaseController) InStatesBlocker(blockStates []string) (inStates bool) {
	// Certi controller non devono subire la cancellazione degli stati
	// perchè magari hanno logiche particolari o lo gestiscono a loro modo
	for _, state := range c.Player.States {
		// Verifico se non fa parte dello stesso padre e che lo stato non sia completato
		if !*state.Completed && (state.Father == 0 || state.Father != c.State.Father) {
			for _, blockState := range blockStates {
				if helpers.Trans(c.Player.Language.Slug, state.Controller) == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("route.%s", blockState)) {
					msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "valodator.controller.blocked"))
					_, err := services.SendMessage(msg)
					if err != nil {
						panic(err)
					}

					return true
				}
			}
		}
	}

	return false
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
