package controllers

//
// import (
// 	"encoding/json"
// 	"log"
// 	"os"
// 	"strings"
// 	"time"
//
// 	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
// 	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
// 	"bitbucket.org/no-name-game/nn-telegram/app/providers"
// 	"bitbucket.org/no-name-game/nn-telegram/services"
// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
// )
//
// // Writer: reloonfire
// // Starting on: 17/01/2020
// // Project: no-name-game
//
// // Crafting:
// // Craft Base effettuati dal player
//
// //====================================
// // CraftingController
// //====================================
// type CraftingV2Controller struct {
// 	BaseController
// 	Payload struct {
// 		Crafted   nnsdk.Crafted // Item da craftare
// 		Resources map[uint]int  // Materiali necessari
// 	}
// }
//
// //====================================
// // Handle
// //====================================
// func (c *CraftingV2Controller) Handle(update tgbotapi.Update) {
// 	// Current Controller instance
// 	var err error
// 	var isNewState bool
// 	c.RouteName, c.Update, c.Message = "route.crafting", update, update.Message
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
//
// 		log.Println("STAGE: ", c.State.Stage)
//
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
// func (c *CraftingV2Controller) Validator() (hasErrors bool) {
// 	var err error
// 	c.Validation.Message = helpers.Trans("validationMessage")
//
// 	switch c.State.Stage {
// 	case 0:
// 		if strings.Contains(c.Message.Text, helpers.Trans("crafting.craft")) {
// 			c.Payload.Crafted, err = providers.GetCraftedByName(strings.Split(c.Message.Text, ": ")[1])
// 			if err != nil {
// 				// Item non esiste
// 				c.Validation.Message = helpers.Trans("crafting.item_does_not_exist")
// 				return true
// 			}
// 			c.State.Stage = 1
// 			return false
// 		}
// 	case 1:
// 		log.Println("VALIDATION 1")
// 		if c.Message.Text == helpers.Trans("yep") {
// 			i, err := providers.GetPlayerInventory(helpers.Player)
// 			if err != nil {
// 				services.ErrorHandler("can't retrieve Inventory", err)
// 			}
// 			// Check for items in inventory
// 			if !helpers.CheckForItems(helpers.InventoryToMap(i), c.Payload.Resources) {
//
// 				c.Validation.Message = helpers.Trans("crafting.no_resource_to_craft")
// 				return true
// 			}
// 			log.Println("YES 2: PASSED")
// 			c.State.Stage = 2
// 			return false
// 		} /*else if c.Message.Text == helpers.Trans("crafting.craft_different_item") {
// 			log.Println("CHANGE CRAFT 1: PASSED")
// 			c.State.Stage = 0
// 			return false
// 		// PERCHE' LO STAGE TORNA A 1???
// 		}*/
// 	case 2:
// 		log.Println("Validation 2")
// 		// Verifico se ha finito il crafting
// 		c.Validation.Message = helpers.Trans("crafting.wait", c.State.FinishAt.Format("15:04:05"))
// 		if time.Now().After(c.State.FinishAt) {
// 			log.Println("Time After 2: PASSED")
// 			c.State.Stage = 3
// 			return false
// 		}
// 	}
//
// 	return true
// }
//
// //====================================
// // Stage  0 What -> 1 - Check Resources -> 2 - Confirm -> 3 - Craft
// //====================================
// func (c *CraftingV2Controller) Stage() {
// 	var err error
//
// 	log.Println("STAGE CHECK:", c.State.Stage)
//
// 	switch c.State.Stage {
// 	case 0:
//
// 		// DEBUG MODE
// 		if os.Getenv("APP_DEBUG") != "false" {
// 			// ASSEGNA 4x OGNI RISORSA AL PLAYER
// 			items, _ := providers.GetAllItems()
// 			for _, item := range items {
// 				_, err := providers.AddResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
// 					ItemID:   item.ID,
// 					Quantity: 4,
// 				})
// 				if err != nil {
// 					services.ErrorHandler("Cant add resource to player inventory", err)
// 				}
// 			}
//
// 		}
//
// 		// =============
// 		// SELEZIONE ITEM
// 		// =============
//
// 		// Lista oggetti craftabili
// 		craftableItems, err := providers.GetAllCraftableItems()
//
// 		if err != nil {
// 			services.ErrorHandler("Can't retrieve Craftable Items", err)
// 		}
// 		// Genero messaggio con tastiera SOLO < R
// 		msg := helpers.ListItemsFilteredBy(craftableItems, 4)
//
// 		services.SendMessage(msg)
// 	case 1:
//
// 		// =============
// 		// CONFERMA ITEM
// 		// =============
//
// 		// Inserisco nel payload la recipelist per avere accesso più facile ad essa
// 		helpers.UnmarshalPayload(c.Payload.Crafted.Recipe.RecipeList, &c.Payload.Resources)
//
// 		// ListRecipe() genera una string contenente gli oggetti necessari al crafting
// 		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("crafting.you_need", c.Payload.Crafted.Item.Name, helpers.ListRecipe(c.Payload.Resources)))
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("yep"))), tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears"))))
// 		msg.ParseMode = "HTML"
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		payloadUpdated, _ := json.Marshal(c.Payload)
// 		c.State.Payload = string(payloadUpdated)
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
//
// 	case 2:
//
// 		var msg tgbotapi.MessageConfig
//
// 		// =============
// 		// START TIMER
// 		// =============
//
// 		// Se il player possiede gli item necessari
//
// 		c.State.FinishAt = helpers.GetEndTime(0, 0, int(c.Payload.Crafted.Recipe.WaitingTime))
// 		c.State.ToNotify = helpers.SetTrue()
//
// 		msg = services.NewMessage(helpers.Player.ChatID, helpers.Trans("crafting.wait", c.State.FinishAt.Format("15:04:05")))
//
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
// 	case 3:
// 		// Aggiungo item all'inventario e rimuovo item
// 		_, err := providers.AddResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
// 			ItemID:   c.Payload.Crafted.ID,
// 			Quantity: 1,
// 		})
// 		if err != nil {
// 			services.ErrorHandler("Cant add resource to player inventory", err)
// 		}
//
// 		// Rimuovo risorse usate al player
// 		for k, q := range c.Payload.Resources {
// 			_, err := providers.RemoveResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
// 				ItemID:   k,
// 				Quantity: q,
// 			})
//
// 			if err != nil {
// 				services.ErrorHandler("Cant remove resource to player inventory", err)
// 			}
// 		}
//
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.craft_completed")+"\n\n"+c.Payload.Crafted.Item.Name)
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
// 		helpers.FinishAndCompleteState(c.State, helpers.Player)
// 		//====================================
//
// 	}
// }
