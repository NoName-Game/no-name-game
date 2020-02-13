package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Crafting:
// Craft Base effettuati dal player

// ====================================
// CraftingController
// ====================================
type CraftingV2Controller struct {
	BaseController
	Payload struct {
		Item      nnsdk.Item   // Item da craftare
		Resources map[uint]int // Materiali necessari
	}
}

// ====================================
// Handle
// ====================================
func (c *CraftingV2Controller) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	c.Controller = "route.crafting"
	c.Player = player
	c.Update = update
	c.Message = update.Message

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// Validate
	var hasError bool
	hasError, err = c.Validator()
	if err != nil {
		panic(err)
	}

	// Se ritornano degli errori
	if hasError == true {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
		_, err = services.SendMessage(validatorMsg)
		if err != nil {
			panic(err)
		}

		return
	}

	// Ok! Run!
	err = c.Stage()
	if err != nil {
		panic(err)
	}

	// Aggiorno stato finale
	_, err = providers.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico se lo stato è completato chiudo
	if *c.State.Completed == true {
		_, err = providers.DeletePlayerState(c.State) // Delete
		if err != nil {
			panic(err)
		}

		err = helpers.DelRedisState(player)
		if err != nil {
			panic(err)
		}
	}

	return
}

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
