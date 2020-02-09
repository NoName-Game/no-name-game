package controllers

//
// import (
// 	"os"
//
// 	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
// 	"bitbucket.org/no-name-game/nn-telegram/services"
// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
// )
//
// type BreakerController struct {
// 	Update    tgbotapi.Update
// 	Message   *tgbotapi.Message
// 	RouteName string
// }
//
// // Back delete only redis state, but not delete state stored in DB.
// type BackController BreakerController
//
// //====================================
// // Handle
// //====================================
// func (c *BackController) Handle(update tgbotapi.Update) {
// 	// Current Controller instance
// 	c.RouteName = "route.breaker.back"
//
// 	// Delete redis state
// 	helpers.DelRedisState(helpers.Player)
//
// 	message := update.Message
// 	if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
// 		msg := services.NewMessage(message.Chat.ID, "***************************\nDEBUG: DELETE REDIS STATE.\n***************************\n")
// 		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
// 		services.SendMessage(msg)
// 	}
//
// 	// Call menu controller
// 	new(MenuController).Handle(update)
// }
//
// // Clears - Delete redist state and remove row from DB.
// type ClearsController BreakerController
//
// //====================================
// // Handle
// //====================================
// func (c *ClearsController) Handle(update tgbotapi.Update) {
// 	// Current Controller instance
// 	c.RouteName = "route.breaker.clears"
//
// 	helpers.DeleteRedisAndDbState(helpers.Player)
//
// 	message := update.Message
// 	if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
// 		msg := services.NewMessage(message.Chat.ID, "***************************\nDEBUG: DELETE DB AND REDIS STATE.\n***************************\n")
// 		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
// 		services.SendMessage(msg)
// 	}
//
// 	// Call menu controller
// 	new(MenuController).Handle(update)
// }
