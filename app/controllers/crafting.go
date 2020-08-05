package controllers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
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
		Item      *pb.Item         // Item da craftare
		Resources map[uint32]int32 // Materiali necessari
	}
}

// ====================================
// Handle
// ====================================
func (c *CraftingController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.crafting",
		c.Payload,
		[]string{},
		player,
		update,
	) {
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.CurrentState.Payload, &c.Payload)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(1, &ShipController{}) {
			return
		}
	}

	// Validate
	var hasError bool
	hasError, err = c.Validator()
	if err != nil {
		panic(err)
	}

	// Se ritornano degli errori
	if hasError {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
		validatorMsg.ReplyMarkup = c.Validation.ReplyKeyboard

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
	c.CurrentState.Payload = string(payloadUpdated)

	rUpdatePlayerState, err := services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
		PlayerState: c.CurrentState,
	})
	if err != nil {
		panic(err)
	}
	c.CurrentState = rUpdatePlayerState.GetPlayerState()

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *CraftingController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		),
	)

	switch c.CurrentState.Stage {
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
		rGetAllItems, err := services.NnSDK.GetAllItems(helpers.NewContext(1), &pb.GetAllItemsRequest{})
		if err != nil {
			return true, err
		}

		// Recupero nome item che il player vuole craftare
		playerChoiche := strings.Split(c.Update.Message.Text, " (")[0]

		var itemExists bool
		for _, item := range rGetAllItems.GetItems() {
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
			rGetPlayerResources, err := services.NnSDK.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
				PlayerID: c.Player.GetID(),
			})
			if err != nil {
				return false, err
			}

			// Ciclo gli elementi di cui devo verificare la presenza
			for resourceID, quantity := range c.Payload.Resources {
				var haveResource bool

				// Ciclo inventario del player
				for _, inventory := range rGetPlayerResources.GetPlayerInventory() {
					if inventory.Resource.ID == resourceID && inventory.Quantity >= quantity {
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
		finishAt, err := ptypes.Timestamp(c.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"crafting.wait",
			finishAt.Format("15:04:05"),
		)

		// Aggiungo anche abbandona
		c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.continue"),
				),
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
				),
			),
		)

		// Verifico se ha finito il crafting
		if time.Now().After(finishAt) {
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
	switch c.CurrentState.Stage {

	// In questo stage invio al player le tipologie di crafting possibili
	case 0:
		rGetAllItemCategories, err := services.NnSDK.GetAllItemCategories(helpers.NewContext(1), &pb.GetAllItemCategoriesRequest{})
		if err != nil {
			return err
		}

		// Creo messaggio
		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.type"))

		var keyboardRow [][]tgbotapi.KeyboardButton
		for _, category := range rGetAllItemCategories.GetItemCategories() {
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
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
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
		c.CurrentState.Stage = 1

	// In questo stage recuperiamo la lista dei ITEMS, appartenenti alla categoria scelta
	// che possono essere anche craftati dal player
	case 1:
		// Recupero tutte le categorie degli items e ciclo per trovare quella voluta del player
		rGetAllItemCategories, err := services.NnSDK.GetAllItemCategories(helpers.NewContext(1), &pb.GetAllItemCategoriesRequest{})
		if err != nil {
			return err
		}

		var chosenCategory *pb.ItemCategory
		for _, category := range rGetAllItemCategories.GetItemCategories() {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("crafting.categories.%s", category.Slug)) {
				chosenCategory = category
			}
		}

		// Lista oggetti craftabili
		rGetItemByCategoryID, err := services.NnSDK.GetItemsByCategoryID(helpers.NewContext(1), &pb.GetItemsByCategoryIDRequest{
			CategoryID: chosenCategory.ID,
		})
		if err != nil {
			return err
		}

		// Recupero tutti gli items del player
		rGetPlayerItems, err := services.NnSDK.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		// Creo messaggio
		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.what"))

		var keyboardRow [][]tgbotapi.KeyboardButton
		for _, item := range rGetItemByCategoryID.GetItems() {
			// Recupero quantità del player per quest'item
			var playerQuantity int32
			for _, playerItem := range rGetPlayerItems.GetPlayerInventory() {
				if playerItem.Item.ID == item.ID {
					playerQuantity = playerItem.Quantity
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
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
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
		c.CurrentState.Stage = 2

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
					helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
				),
			),
		)
		msg.ParseMode = "HTML"
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3

	// In questo stage mi aspetto che l'utente abbia confermato e se così fosse
	// procedo con il rimuovere le risorse associate e notificargli l'attesa per il crafting
	case 3:
		// Rimuovo risorse usate al player
		for resourceID, quantity := range c.Payload.Resources {
			_, err := services.NnSDK.ManagePlayerInventory(helpers.NewContext(1), &pb.ManagePlayerInventoryRequest{
				PlayerID: c.Player.GetID(),
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

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorna stato
		// c.CurrentState.FinishAt = endTime
		c.CurrentState.ToNotify = true
		c.CurrentState.Stage = 4
		c.Breaker.ToMenu = true

	// In questo stage il player ha completato correttamente il crafting, quindi
	// proseguo con l'assegnarli l'item e concludo
	case 4:
		// Aggiungo item all'inventario
		_, err := services.NnSDK.ManagePlayerInventory(helpers.NewContext(1), &pb.ManagePlayerInventoryRequest{
			PlayerID: c.Player.GetID(),
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
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
