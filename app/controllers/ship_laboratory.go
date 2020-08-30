package controllers

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipLaboratoryController
// Ogni player ha la possibilità di craftare al player degli item
// che possono essere usati in diversi modo, es. per recuperare vita
// o per ripristinare determinate cose
// ====================================
type ShipLaboratoryController struct {
	BaseController
	Payload struct {
		Item      *pb.Item         // Item da craftare
		Resources map[uint32]int32 // Materiali necessari
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipLaboratoryController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller: "route.ship.laboratory",
		ControllerBack: ControllerBack{
			To:        &ShipController{},
			FromStage: 1,
		},
		Payload: c.Payload,
	}) {
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)

	// Validate
	var hasError bool
	if hasError = c.Validator(); hasError {
		c.Validate()
		return
	}

	// Ok! Run!
	if err = c.Stage(); err != nil {
		panic(err)
	}

	// Completo progressione
	if err = c.Completing(c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *ShipLaboratoryController) Validator() (hasErrors bool) {
	var err error
	switch c.PlayerData.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	// In questo stage verifico se mi è stata passata una categoria che esiste realmente
	case 1:
		// category, err = providers.FindItemCategoryByName()
		if !helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.categories.medical"),
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.categories.ship_support"),
		}) {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			return true
		}

		return false
	// In questo stage è necessario verificare se il player ha passato un item che eiste realmente
	case 2:
		// Recupero tutte gli items e ciclo per trovare quello voluta del player
		var rGetAllItems *pb.GetAllItemsResponse
		rGetAllItems, err = services.NnSDK.GetAllItems(helpers.NewContext(1), &pb.GetAllItemsRequest{})
		if err != nil {
			return true
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
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.laboratory.item_does_not_exist")

			return true
		}

		return false
	// In questo stage è necessario che venga validato se il player ha tutti i
	// materiali necessario al crafting dell'item da lui scelto
	case 3:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "yep") {
			// Verifico se il player ha tutto gli item necessari
			var rGetPlayerResources *pb.GetPlayerResourcesResponse
			rGetPlayerResources, err = services.NnSDK.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
				PlayerID: c.Player.GetID(),
			})
			if err != nil {
				return false
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
					c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.laboratory.no_resource_to_craft")

					return true
				}
			}

			return false
		}

		return true

	// In questo stage verificho che l'utente abbia effettivamente aspettato
	// il tempo di attesa necessario al craft
	case 4:
		var finishAt time.Time
		finishAt, err = ptypes.Timestamp(c.PlayerData.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"ship.laboratory.wait",
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
			return false
		}

		return true
	}

	return true
}

// ====================================
// Stage  0 What -> 1 - Check Resources -> 2 - Confirm -> 3 - Craft
// ====================================
func (c *ShipLaboratoryController) Stage() (err error) {
	switch c.PlayerData.CurrentState.Stage {

	// In questo stage invio al player le tipologie di crafting possibili
	case 0:
		var rGetAllItemCategories *pb.GetAllItemCategoriesResponse
		rGetAllItemCategories, err = services.NnSDK.GetAllItemCategories(helpers.NewContext(1), &pb.GetAllItemCategoriesRequest{})
		if err != nil {
			return err
		}

		laboratoryInfo := fmt.Sprintf("%s\n\n%s",
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.type"),
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.info"),
		)

		// Creo messaggio
		msg := services.NewMessage(c.Player.ChatID, laboratoryInfo)

		var keyboardRow [][]tgbotapi.KeyboardButton
		for _, category := range rGetAllItemCategories.GetItemCategories() {
			row := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ship.laboratory.categories.%s", category.Slug)),
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
		c.PlayerData.CurrentState.Stage = 1

	// In questo stage recuperiamo la lista dei ITEMS, appartenenti alla categoria scelta
	// che possono essere anche craftati dal player
	case 1:
		// Recupero tutte le categorie degli items e ciclo per trovare quella voluta del player
		var rGetAllItemCategories *pb.GetAllItemCategoriesResponse
		rGetAllItemCategories, err = services.NnSDK.GetAllItemCategories(helpers.NewContext(1), &pb.GetAllItemCategoriesRequest{})
		if err != nil {
			return err
		}

		var chosenCategory *pb.ItemCategory
		for _, category := range rGetAllItemCategories.GetItemCategories() {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ship.laboratory.categories.%s", category.Slug)) {
				chosenCategory = category
			}
		}

		// Lista oggetti craftabili
		var rGetItemByCategoryID *pb.GetItemsByCategoryIDResponse
		rGetItemByCategoryID, err = services.NnSDK.GetItemsByCategoryID(helpers.NewContext(1), &pb.GetItemsByCategoryIDRequest{
			CategoryID: chosenCategory.ID,
		})
		if err != nil {
			return err
		}

		// Recupero tutti gli items del player
		var rGetPlayerItems *pb.GetPlayerItemsResponse
		rGetPlayerItems, err = services.NnSDK.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		// Creo messaggio
		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "ship.laboratory.what"))

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
		c.PlayerData.CurrentState.Stage = 2

	// In questo stage riepilogo le risorse necessarie e
	// chiedo al conferma al player se continuare il crafting dell'item
	case 2:
		// Inserisco nel payload la recipelist per avere accesso più facile ad essa
		helpers.UnmarshalPayload(c.Payload.Item.Recipe.RecipeList, &c.Payload.Resources)

		// Genero string contenente le risorse richieste per il craft
		var itemsRecipeList string
		for resourceID, value := range c.Payload.Resources {
			var rGetResourceByID *pb.GetResourceByIDResponse
			rGetResourceByID, err = services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: resourceID,
			})
			if err != nil {
				return err
			}

			itemsRecipeList += fmt.Sprintf("- *%v* x %s (%s)\n",
				value,
				rGetResourceByID.GetResource().GetName(),
				rGetResourceByID.GetResource().GetRarity().GetSlug(),
			)
		}

		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"ship.laboratory.you_need",
				helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
				itemsRecipeList,
			),
		)
		msg.ParseMode = "markdown"
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
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.PlayerData.CurrentState.Stage = 3

	// In questo stage mi aspetto che l'utente abbia confermato e se così fosse
	// procedo con il rimuovere le risorse associate e notificargli l'attesa per il crafting
	case 3:
		// TODO: logica da spostare su WS
		// Rimuovo risorse usate al player
		for resourceID, quantity := range c.Payload.Resources {
			_, err = services.NnSDK.ManagePlayerInventory(helpers.NewContext(1), &pb.ManagePlayerInventoryRequest{
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
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.wait", endTime.Format("15:04:05")),
		)
		msg.ParseMode = "markdown"
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorna stato
		// c.PlayerData.CurrentState.FinishAt = endTime
		c.PlayerData.CurrentState.ToNotify = true
		c.PlayerData.CurrentState.Stage = 4
		c.ForceBackTo = true

	// In questo stage il player ha completato correttamente il crafting, quindi
	// proseguo con l'assegnarli l'item e concludo
	case 4:
		// TODO: logica da spostare su WS
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
				"ship.laboratory.craft_completed",
				helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
			),
		)
		msg.ParseMode = "markdown"
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
		c.PlayerData.CurrentState.Completed = true
	}

	return
}
