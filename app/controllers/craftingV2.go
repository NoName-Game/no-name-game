package controllers

import (
	"encoding/json"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Crafting:
// Craft Base effettuati dal player
//TODO: andrebbe dato uno costo se comprato dall'npc una categoria per
// per differenziare e una descrizione dell'oggetto

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

// ====================================
// Validator
// ====================================
func (c *CraftingV2Controller) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	switch c.State.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	// In questo stage è necessario verificare se il player ha passato un item che eiste realmente
	case 1:
		if strings.Contains(c.Message.Text, helpers.Trans(c.Player.Language.Slug, "crafting.craft")) {
			c.Payload.Item, err = providers.GetItemByName(strings.Split(c.Message.Text, ": ")[1])
			// Item non esiste
			if err != nil {
				c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "crafting.item_does_not_exist")

				return true, err
			}

			return false, err
		}

		// // In questo stage è necessario che venga validato se il player ha tutti i
		// // materiali necessario al crafting dell'item da lui scelto
		// case 2:
		// 	if c.Message.Text != helpers.Trans(c.Player.Language.Slug, "yep") {
		//
		// 		return false
		// 	} /*else if c.Message.Text == helpers.Trans("crafting.craft_different_item") {
		// 		log.Println("CHANGE CRAFT 1: PASSED")
		// 		c.State.Stage = 0
		// 		return false
		// 	// PERCHE' LO STAGE TORNA A 1???
		// 	}*/
		// case 3:
		// 	log.Println("Validation 2")
		//
		// 	// Verifico se ha finito il crafting
		// 	c.Validation.Message = helpers.Trans("crafting.wait", c.State.FinishAt.Format("15:04:05"))
		// 	if time.Now().After(c.State.FinishAt) {
		// 		log.Println("Time After 2: PASSED")
		// 		c.State.Stage = 3
		// 		return false
		// 	}
	}

	return true, err
}

// ====================================
// Stage  0 What -> 1 - Check Resources -> 2 - Confirm -> 3 - Craft
// ====================================
func (c *CraftingV2Controller) Stage() (err error) {
	switch c.State.Stage {

	// In questo stage recuperiamo la lista dei ITEMS, ovvero
	// quegli oggetti che possono essere anche craftati dal player
	case 0:
		// Lista oggetti craftabili
		craftableItems, err := providers.GetAllItems()
		if err != nil {
			return err
		}

		// Creo messaggio
		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.type"))

		var keyboardRow [][]tgbotapi.KeyboardButton
		for _, item := range craftableItems {
			row := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "crafting.craft") + item.Name),
			)
			keyboardRow = append(keyboardRow, row)
		}

		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRow,
			ResizeKeyboard: false,
		}

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Avanzo di stage
		c.State.Stage = 1

	// In questo stage riepilogo le risorse necessarie e
	// chiedo al conferma al player se continuare il crafting dell'item
	case 1:
		// Inserisco nel payload la recipelist per avere accesso più facile ad essa
		helpers.UnmarshalPayload(c.Payload.Item.Recipe.RecipeList, &c.Payload.Resources)

		// ListRecipe() genera una string contenente gli oggetti necessari al crafting
		var itemsRecipeList string
		itemsRecipeList, err = helpers.ListRecipe(c.Payload.Resources)
		if err != nil {
			return err
		}

		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "crafting.you_need", c.Payload.Item.Name, itemsRecipeList),
		)

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "yep"),
				),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
				),
			),
		)
		msg.ParseMode = "HTML"
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)

		// case 2:
		//
		// 	var msg tgbotapi.MessageConfig
		//
		// 	// =============
		// 	// START TIMER
		// 	// =============
		//
		// 	// Se il player possiede gli item necessari
		//
		// 	c.State.FinishAt = helpers.GetEndTime(0, 0, int(c.Payload.Crafted.Recipe.WaitingTime))
		// 	c.State.ToNotify = helpers.SetTrue()
		//
		// 	msg = services.NewMessage(helpers.Player.ChatID, helpers.Trans("crafting.wait", c.State.FinishAt.Format("15:04:05")))
		//
		// 	services.SendMessage(msg)
		//
		// 	// Aggiorno stato
		// 	c.State, err = providers.UpdatePlayerState(c.State)
		// 	if err != nil {
		// 		services.ErrorHandler("Cant update player", err)
		// 	}
		// case 3:
		// 	// Aggiungo item all'inventario e rimuovo item
		// 	_, err := providers.AddResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
		// 		ItemID:   c.Payload.Crafted.ID,
		// 		Quantity: 1,
		// 	})
		// 	if err != nil {
		// 		services.ErrorHandler("Cant add resource to player inventory", err)
		// 	}
		//
		// 	// Rimuovo risorse usate al player
		// 	for k, q := range c.Payload.Resources {
		// 		_, err := providers.RemoveResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
		// 			ItemID:   k,
		// 			Quantity: q,
		// 		})
		//
		// 		if err != nil {
		// 			services.ErrorHandler("Cant remove resource to player inventory", err)
		// 		}
		// 	}
		//
		// 	// Invio messaggio
		// 	msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.craft_completed")+"\n\n"+c.Payload.Crafted.Item.Name)
		// 	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		// 		tgbotapi.NewKeyboardButtonRow(
		// 			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
		// 		),
		// 	)
		// 	services.SendMessage(msg)
		//
		// 	//====================================
		// 	// COMPLETE!
		// 	//====================================
		// 	helpers.FinishAndCompleteState(c.State, helpers.Player)
		// 	//====================================
		//
	}

	return
}
