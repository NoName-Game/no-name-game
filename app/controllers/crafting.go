package controllers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// CraftingController
// Ogni player ha la possibilità di craftare al player degli item
// che possono essere usati in diversi modo, es. per recuperare vita
// o per ripristinare determinate cose
// ====================================
type CraftingController struct {
	BaseController
	Payload struct {
		Item      nnsdk.Item   // Item da craftare
		Resources map[uint]int // Materiali necessari
	}
}

// ====================================
// Handle
// ====================================
func (c *CraftingController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	var playerStateProvider providers.PlayerStateProvider

	c.Controller = "route.crafting"
	c.Player = player
	c.Update = update

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
		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
		validatorMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
				),
			),
		)

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
	payloadUpdated, _ := json.Marshal(c.Payload)
	c.State.Payload = string(payloadUpdated)
	c.State, err = playerStateProvider.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}

	return
}

// ====================================
// Validator
// ====================================
func (c *CraftingController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	var itemProvider providers.ItemProvider
	var playerProvider providers.PlayerProvider

	switch c.State.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	// In questo stage verifico se mi è stata passata una categoria che esiste realmente
	case 1:
		// category, err = providers.FindItemCategoryByName()
		if !helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "crafting.categories.medical"),
			helpers.Trans(c.Player.Language.Slug, "crafting.categories.ship_support"),
		}) {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			return true, err
		}

		return false, err
	// In questo stage è necessario verificare se il player ha passato un item che eiste realmente
	case 2:
		// Recupero tutte gli items e ciclo per trovare quello voluta del player
		var items nnsdk.Items
		items, err = itemProvider.GetAllItems()
		if err != nil {
			return true, err
		}

		// Recupero nome item che il player vuole craftare
		playerChoiche := strings.Split(c.Update.Message.Text, " (")[0]

		var itemExists bool
		for _, item := range items {
			if playerChoiche == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", item.Slug)) {
				itemExists = true
				c.Payload.Item = item
			}
		}

		if !itemExists {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "crafting.item_does_not_exist")

			return true, err
		}

		return false, err
	// In questo stage è necessario che venga validato se il player ha tutti i
	// materiali necessario al crafting dell'item da lui scelto
	case 3:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "yep") {
			// Verifico se il player ha tutto gli item necessari
			var playerInventory nnsdk.PlayerInventories
			playerInventory, _ = playerProvider.GetPlayerResources(c.Player.ID)

			// Ciclo gli elementi di cui devo verificare la presenza
			for resourceID, quantity := range c.Payload.Resources {
				var haveResource bool

				// Ciclo inventario del player
				for _, inventory := range playerInventory {
					if inventory.Resource.ID == resourceID && *inventory.Quantity >= quantity {
						haveResource = true
					}
				}

				// Basta che anche solo una vola ritorni false per far fallire
				// il controllo in quanti per continuare di stage il player
				// deve possedere TUTTI gli elementi
				if !haveResource {
					c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "crafting.no_resource_to_craft")

					return true, err
				}
			}

			return false, err
		}

		return true, err

	// In questo stage verificho che l'utente abbia effettivamente aspettato
	// il tempo di attesa necessario al craft
	case 4:
		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"crafting.wait",
			c.State.FinishAt.Format("15:04:05"),
		)

		// Verifico se ha finito il crafting
		if time.Now().After(c.State.FinishAt) {
			return false, err
		}

		return true, err
	}

	return true, err
}

// ====================================
// Stage  0 What -> 1 - Check Resources -> 2 - Confirm -> 3 - Craft
// ====================================
func (c *CraftingController) Stage() (err error) {
	var itemCategoryProvider providers.ItemCategoryProvider
	var itemProvider providers.ItemProvider
	var playerProvider providers.PlayerProvider
	var playerInventoryProvider providers.PlayerProvider

	switch c.State.Stage {

	// In questo stage invio al player le tipologie di crafting possibili
	case 0:
		var itemCategories nnsdk.ItemCategories
		itemCategories, err = itemCategoryProvider.GetAllItemCategories()
		if err != nil {
			return err
		}

		// Creo messaggio
		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.type"))

		var keyboardRow [][]tgbotapi.KeyboardButton
		for _, category := range itemCategories {
			row := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("crafting.categories.%s", category.Slug)),
				),
			)
			keyboardRow = append(keyboardRow, row)
		}

		// Aggiungo bottone cancella
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
			),
		))

		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRow,
			ResizeKeyboard: true,
		}

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Avanzo di stage
		c.State.Stage = 1

	// In questo stage recuperiamo la lista dei ITEMS, appartenenti alla categoria scelta
	// che possono essere anche craftati dal player
	case 1:
		// Recupero tutte le categorie degli items e ciclo per trovare quella voluta del player
		var itemCategories nnsdk.ItemCategories
		itemCategories, err = itemCategoryProvider.GetAllItemCategories()
		if err != nil {
			return err
		}

		var chosenCategory nnsdk.ItemCategory
		for _, category := range itemCategories {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("crafting.categories.%s", category.Slug)) {
				chosenCategory = category
			}
		}

		// Lista oggetti craftabili
		var craftableItems nnsdk.Items
		craftableItems, err = itemProvider.GetItemByCategoryID(chosenCategory.ID)
		if err != nil {
			return err
		}

		// Recupero tutti gli items del player
		var playerInventoryItems nnsdk.PlayerInventories
		playerInventoryItems, err = playerProvider.GetPlayerItems(c.Player.ID)
		if err != nil {
			panic(err)
		}

		// Creo messaggio
		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.what"))

		var keyboardRow [][]tgbotapi.KeyboardButton
		for _, item := range craftableItems {
			// Recupero quantità del player per quest'item
			var playerQuantity int
			for _, playerItem := range playerInventoryItems {
				if playerItem.Item.ID == item.ID {
					playerQuantity = *playerItem.Quantity
				}
			}

			row := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					fmt.Sprintf(
						"%s (%v)",
						helpers.Trans(c.Player.Language.Slug, "items."+item.Slug),
						playerQuantity,
					),
				),
			)
			keyboardRow = append(keyboardRow, row)
		}

		// Aggiungo bottone cancella
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
			),
		))

		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRow,
			ResizeKeyboard: true,
		}

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Avanzo di stage
		c.State.Stage = 2

	// In questo stage riepilogo le risorse necessarie e
	// chiedo al conferma al player se continuare il crafting dell'item
	case 2:
		// Inserisco nel payload la recipelist per avere accesso più facile ad essa
		helpers.UnmarshalPayload(c.Payload.Item.Recipe.RecipeList, &c.Payload.Resources)

		// ListRecipe() genera una string contenente gli oggetti necessari al crafting
		var itemsRecipeList string
		itemsRecipeList, err = helpers.ListRecipe(c.Payload.Resources)
		if err != nil {
			return err
		}

		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"crafting.you_need",
				helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
				itemsRecipeList,
			),
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
		c.State.Stage = 3

	// In questo stage mi aspetto che l'utente abbia confermato e se così fosse
	// procedo con il rimuovere le risorse associate e notificargli l'attesa per il crafting
	case 3:
		// Rimuovo risorse usate al player
		for resourceID, quantity := range c.Payload.Resources {
			err = playerInventoryProvider.ManagePlayerInventory(c.Player.ID, nnsdk.ManageInventoryRequest{
				ItemID:   resourceID,
				ItemType: "resources",
				Quantity: -quantity,
			})
			if err != nil {
				return err
			}
		}

		// Definisco endtime per il crafting
		endTime := helpers.GetEndTime(0, 0, int(c.Payload.Item.Recipe.WaitingTime))

		// Notifico
		var msg tgbotapi.MessageConfig
		msg = services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "crafting.wait", endTime.Format("15:04:05")),
		)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorna stato
		c.State.FinishAt = endTime
		c.State.ToNotify = helpers.SetTrue()
		c.State.Stage = 4

	// In questo stage il player ha completato correttamente il crafting, quindi
	// proseguo con l'assegnarli l'item e concludo
	case 4:
		// Aggiungo item all'inventario
		err = playerInventoryProvider.ManagePlayerInventory(c.Player.ID, nnsdk.ManageInventoryRequest{
			ItemID:   c.Payload.Item.ID,
			ItemType: "items",
			Quantity: 1,
		})
		if err != nil {
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(
				c.Player.Language.Slug,
				"crafting.craft_completed",
				helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
			),
		)

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.State.Completed = helpers.SetTrue()
	}

	return
}
