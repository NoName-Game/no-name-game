package controllers

//
// import (
// 	"time"
//
// 	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
// 	"bitbucket.org/no-name-game/nn-telegram/app/providers"
// 	"bitbucket.org/no-name-game/nn-telegram/services"
// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
// )
//
// //====================================
// // DeathController
// //====================================
// type DeathController struct {
// 	BaseController
// 	Payload struct {
// 		Type    string
// 		EquipID uint
// 	}
// }
//
// //====================================
// // Handle
// //====================================
// func (c *DeathController) Handle(update tgbotapi.Update) {
// 	// Current Controller instance
// 	var err error
// 	var isNewState bool
// 	c.RouteName, c.Update, c.Message = "route.death", update, update.Message
//
// 	// Check current state for this routes
// 	c.State, isNewState = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)
//
// 	// Set and load payload
// 	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)
//
// 	// It's first message
// 	if isNewState {
// 		c.Stage()
// 		return
// 	}
//
// 	// Go to validator
// 	if !c.Validator() {
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
//
// 		// Ok! Run!
// 		c.Stage()
// 		return
// 	}
//
// 	// Validator goes errors
// 	validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
// 	services.SendMessage(validatorMsg)
// 	return
// }
//
// //====================================
// // Validator
// //====================================
// func (c *DeathController) Validator() (hasErrors bool) {
// 	c.Validation.Message = helpers.Trans("playerDie", c.State.FinishAt.Format("3:04PM"))
//
// 	switch c.State.Stage {
// 	case 0:
// 		return false
// 	case 1:
// 		if time.Now().After(c.State.FinishAt) {
// 			return false
// 		}
// 	}
//
// 	return true
// }
//
// //====================================
// // Stage
// //====================================
// func (c *DeathController) Stage() {
// 	var err error
//
// 	switch c.State.Stage {
// 	case 0:
//
// 		// Invio messaggio
// 		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("playerDie", c.State.FinishAt.Format("04:05")))
// 		msg.ParseMode = "HTML"
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		c.State.Stage = 1
// 		c.State.ToNotify = helpers.SetTrue()
// 		c.State.FinishAt = time.Now().Add((time.Hour * time.Duration(12)))
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
// 	case 1:
// 		stats, err := providers.GetPlayerStats(helpers.Player)
// 		if err != nil {
// 			services.ErrorHandler("Cant retrieve stats", err)
// 		}
//
// 		*stats.LifePoint = 100 + stats.Level*10
//
// 		_, err = providers.UpdatePlayerStats(stats)
// 		if err != nil {
// 			services.ErrorHandler("Cant update stats", err)
// 		}
//
// 		//====================================
// 		// COMPLETE!
// 		//====================================
// 		helpers.FinishAndCompleteState(c.State, helpers.Player)
// 		//====================================
// 	}
// }
