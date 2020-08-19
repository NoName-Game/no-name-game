package controllers

//
// import (
// 	"encoding/json"
// 	"strings"
//
// 	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
// 	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
// 	"bitbucket.org/no-name-game/nn-telegram/app/providers"
// 	"bitbucket.org/no-name-game/nn-telegram/services"
// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
// )
//
// //====================================
// // InventoryDestroyController
// //====================================
// type InventoryDestroyController struct {
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
// func (c *InventoryDestroyController) Handle(update tgbotapi.Update) {
// 	// Current Controller instance
// 	var err error
// 	var isNewState bool
// 	c.RouteName, c.Update, c.Message = "route.inventory.destroy", update, update.Message
//
// 	// Check current state for this routes
// 	c.PlayerData.CurrentState, isNewState = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)
//
// 	// Set and load payload
// 	helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)
//
// 	// It's first message
// 	if isNewState {
// 		c.Stage()
// 		return
// 	}
//
// 	// Go to validator
// 	if !c.Validator() {
// 		c.PlayerData.CurrentState, err = providers.UpdatePlayerState(c.PlayerData.CurrentState)
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
// func (c *InventoryDestroyController) Validator() (hasErrors bool) {
// 	c.Validation.Message = helpers.Trans("validationMessage")
//
// 	switch c.PlayerData.CurrentState.Stage {
// 	case 0:
// 		if helpers.InArray(c.Message.Text, []string{
// 			helpers.Trans("armors"),
// 			helpers.Trans("weapons"),
// 		}) {
// 			c.PlayerData.CurrentState.Stage = 1
// 			return false
// 		}
// 	case 1:
// 		if strings.Contains(c.Message.Text, helpers.Trans("destroy")) {
// 			c.PlayerData.CurrentState.Stage = 2
// 			return false
// 		}
// 	case 2:
// 		if c.Message.Text == helpers.Trans("confirm") {
// 			c.PlayerData.CurrentState.Stage = 3
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
// func (c *InventoryDestroyController) Stage() {
// 	var err error
//
// 	switch c.PlayerData.CurrentState.Stage {
// 	case 0:
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.destroy.type"))
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("armors")),
// 			),
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("weapons")),
// 			),
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
// 			),
// 		)
// 		services.SendMessage(msg)
// 	case 1:
// 		var keyboardRowCategories [][]tgbotapi.KeyboardButton
// 		switch c.Payload.Type {
// 		case helpers.Trans("armors"):
// 			// Each player armors
// 			for _, armor := range helpers.Player.Armors {
// 				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("destroy") + " " + armor.Name))
// 				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
// 			}
// 		case helpers.Trans("weapons"):
// 			// Each player weapons
// 			for _, weapon := range helpers.Player.Weapons {
// 				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("destroy") + " " + weapon.Name))
// 				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
// 			}
// 		}
//
// 		// Clear and exit
// 		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
// 		))
//
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.destroy.what"))
// 		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
// 			ResizeKeyboard: true,
// 			Keyboard:       keyboardRowCategories,
// 		}
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		c.Payload.Type = c.Message.Text
// 		payloadUpdated, _ := json.Marshal(c.Payload)
// 		c.PlayerData.CurrentState.Payload = string(payloadUpdated)
// 		c.PlayerData.CurrentState, err = providers.UpdatePlayerState(c.PlayerData.CurrentState)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
// 	case 2:
// 		var equipmentName string
//
// 		// Ripulisco messaggio
// 		equipmentName = strings.Split(c.Message.Text, helpers.Trans("destroy")+" ")[1]
//
// 		var equipmentID uint
// 		switch c.Payload.Type {
// 		case helpers.Trans("armors"):
// 			var armor nnsdk.Armor
// 			armor, err := providers.FindArmorByName(equipmentName)
// 			if err != nil {
// 				services.ErrorHandler("Cant find equip armor name", err)
// 			}
//
// 			equipmentID = armor.ID
// 		case helpers.Trans("weapons"):
// 			var weapon nnsdk.Weapon
// 			weapon, err := providers.FindWeaponByName(equipmentName)
// 			if err != nil {
// 				services.ErrorHandler("Cant find equip weapon name", err)
// 			}
//
// 			equipmentID = weapon.ID
// 		}
//
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.destroy.confirm")+"\n\n "+equipmentName)
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("confirm")),
// 			),
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
// 			),
// 		)
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		c.Payload.EquipID = equipmentID
// 		payloadUpdated, _ := json.Marshal(c.Payload)
// 		c.PlayerData.CurrentState.Payload = string(payloadUpdated)
// 		c.PlayerData.CurrentState, err = providers.UpdatePlayerState(c.PlayerData.CurrentState)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
// 	case 3:
// 		switch c.Payload.Type {
// 		case helpers.Trans("armors"):
// 			equipment, err := providers.GetArmorByID(c.Payload.EquipID)
// 			if err != nil {
// 				services.ErrorHandler("Cant find weapon by ID", err)
// 			}
//
// 			_, err = providers.DeleteArmor(equipment)
// 			if err != nil {
// 				services.ErrorHandler("Cant delete armor", err)
// 			}
// 		case helpers.Trans("weapons"):
// 			equipment, err := providers.GetWeaponByID(c.Payload.EquipID)
// 			if err != nil {
// 				services.ErrorHandler("Cant find weapon by ID", err)
// 			}
//
// 			_, err = providers.DeleteWeapon(equipment)
// 			if err != nil {
// 				services.ErrorHandler("Cant delete weapon", err)
// 			}
// 		}
//
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.destroy.completed"))
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 			),
// 		)
// 		services.SendMessage(msg)
//
// 		//====================================
// 		// COMPLETE!
// 		//====================================
// 		helpers.FinishAndCompleteState(c.PlayerData.CurrentState, helpers.Player)
// 		//====================================
//
// 	}
// }
