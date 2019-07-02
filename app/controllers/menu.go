package controllers

import (
	"bitbucket.org/no-name-game/no-name/services"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/provider"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Menu - create a menu with tasks to finish and other messages.
func Menu(update tgbotapi.Update) {

	/* Computer di bordo di reloonfire
	Task in corso:
	Tutorial.
	Crafting: Termina alle ore xx:xx:xx.
	Versione di sviluppo di NoNameGame, tutti i testi potranno cambiare con la release ufficiale.
	*/

	// No state in DB, only show task and a keyboard.

	var tasks string
	var keyboardRows [][]tgbotapi.KeyboardButton

	states, _ := provider.GetPlayerStates(helpers.Player)

	for _, state := range states {
		if *state.ToNotify {
			// If FinishAt is setted "On Going %TASKNAME: Finish at XX:XX:XX"
			stateText := helpers.Trans(state.Function) + helpers.Trans("menu.finishAt", state.FinishAt.Format("15:04:05"))
			tasks += helpers.Trans("menu.onGoing", stateText) + "\n"
		} else {
			tasks += helpers.Trans("menu.onGoing", helpers.Trans(state.Function)) + "\n"
		}
		keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(state.Function)))
		keyboardRows = append(keyboardRows, keyboardRow)
	}

	msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("menu", helpers.Player.Username, tasks))
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard: keyboardRows,
	}
	services.SendMessage(msg)

}
