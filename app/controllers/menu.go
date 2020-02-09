package controllers

//
// import (
// 	"bitbucket.org/no-name-game/nn-telegram/services"
//
// 	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
// )
//
// /*
// Example:
// Computer di bordo di reloonfire
// Task in corso:
// Tutorial.
// Crafting: Termina alle ore xx:xx:xx.
// Versione di sviluppo di NoNameGame, tutti i testi potranno cambiare con la release ufficiale.
// */
//
// type MenuController BaseController
//
// //====================================
// // Handle
// //====================================
// func (c *MenuController) Handle(update tgbotapi.Update) {
// 	// Current Controller instance
// 	c.RouteName = "route.abilityTree"
//
// 	// Keyboard menu
// 	var keyboardMenu = [][]tgbotapi.KeyboardButton{
// 		{tgbotapi.NewKeyboardButton(helpers.Trans("route.mission")), tgbotapi.NewKeyboardButton(helpers.Trans("route.hunting"))},
// 		{tgbotapi.NewKeyboardButton(helpers.Trans("route.inventory"))},
// 		{tgbotapi.NewKeyboardButton(helpers.Trans("route.crafting")), tgbotapi.NewKeyboardButton(helpers.Trans("route.abilityTree"))},
// 	}
//
// 	var tasks string
// 	var keyboardRows [][]tgbotapi.KeyboardButton
// 	msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("menu", helpers.Player.Username, tasks))
// 	msg.ParseMode = "HTML"
//
// 	for _, state := range helpers.Player.States {
// 		if *state.ToNotify {
// 			// If FinishAt is setted "On Going %TASKNAME: Finish at XX:XX:XX"
// 			stateText := helpers.Trans(state.Function) + helpers.Trans("menu.finishAt", state.FinishAt.Format("15:04:05"))
// 			tasks += helpers.Trans("menu.onGoing", stateText) + "\n"
// 		} else {
// 			tasks += helpers.Trans("menu.onGoing", helpers.Trans(state.Function)) + "\n"
// 		}
// 		keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(state.Function)))
// 		keyboardRows = append(keyboardRows, keyboardRow)
// 	}
//
// 	for _, state := range helpers.Player.States {
// 		if state.Function == "route.start" {
// 			msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
// 				Keyboard: keyboardRows,
// 			}
// 			break
// 		} else {
// 			msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
// 				Keyboard: keyboardMenu,
// 			}
// 		}
// 	}
//
// 	// Send recap message
// 	services.SendMessage(msg)
// }
