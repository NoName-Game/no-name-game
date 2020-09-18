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

	// Carico payload
	if err = helpers.GetPayloadController(c.Player.ID, c.CurrentState.Controller, &c.Payload); err != nil {
		panic(err)
	}

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
	if err = c.Completing(&c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *ShipLaboratoryController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
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
		if rGetAllItems, err = services.NnSDK.GetAllItems(helpers.NewContext(1), &pb.GetAllItemsRequest{}); err != nil {
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
			var rLaboratoryCheckHaveResourceForCrafting *pb.LaboratoryCheckHaveResourceForCraftingResponse
			if rLaboratoryCheckHaveResourceForCrafting, err = services.NnSDK.LaboratoryCheckHaveResourceForCrafting(helpers.NewContext(1), &pb.LaboratoryCheckHaveResourceForCraftingRequest{
				PlayerID: c.Player.ID,
				ItemID:   c.Payload.Item.ID,
			}); err != nil {
				panic(err)
			}

			if !rLaboratoryCheckHaveResourceForCrafting.GetHaveResources() {
				c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.laboratory.no_resource_to_craft")
				return true
			}
			return false
		}

		return true

	// In questo stage verificho che l'utente abbia effettivamente aspettato
	// il tempo di attesa necessario al craft
	case 4:
		var rLaboratoryCheckCrafting *pb.LaboratoryCheckCraftingResponse
		if rLaboratoryCheckCrafting, err = services.NnSDK.LaboratoryCheckCrafting(helpers.NewContext(1), &pb.LaboratoryCheckCraftingRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			panic(err)
		}

		// Il crafter sta già portando a terminre un lavoro per questo player
		if !rLaboratoryCheckCrafting.GetFinishCrafting() {
			var finishAt time.Time
			finishAt, err = ptypes.Timestamp(rLaboratoryCheckCrafting.GetCraftingEndTime())
			if err != nil {
				panic(err)
			}

			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"ship.laboratory.wait_validator",
				finishAt.Format("15:04:05"),
			)

			return true
		}

		return false
	}

	return true
}

// ====================================
// Stage  0 What -> 1 - Check Resources -> 2 - Confirm -> 3 - Craft
// ====================================
func (c *ShipLaboratoryController) Stage() (err error) {
	switch c.CurrentState.Stage {

	// In questo stage invio al player le tipologie di crafting possibili
	case 0:
		var rGetAllItemCategories *pb.GetAllItemCategoriesResponse
		if rGetAllItemCategories, err = services.NnSDK.GetAllItemCategories(helpers.NewContext(1), &pb.GetAllItemCategoriesRequest{}); err != nil {
			return err
		}

		// Creo messaggio
		msg := services.NewMessage(c.Player.ChatID, fmt.Sprintf("%s\n\n%s",
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.type"),
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.info"),
		))

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

		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1

	// In questo stage recuperiamo la lista dei ITEMS, appartenenti alla categoria scelta
	// che possono essere anche craftati dal player
	case 1:
		// Recupero tutte le categorie degli items e ciclo per trovare quella voluta del player
		var rGetAllItemCategories *pb.GetAllItemCategoriesResponse
		if rGetAllItemCategories, err = services.NnSDK.GetAllItemCategories(helpers.NewContext(1), &pb.GetAllItemCategoriesRequest{}); err != nil {
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
		if rGetItemByCategoryID, err = services.NnSDK.GetItemsByCategoryID(helpers.NewContext(1), &pb.GetItemsByCategoryIDRequest{
			CategoryID: chosenCategory.ID,
		}); err != nil {
			return err
		}

		// Recupero tutti gli items del player
		var rGetPlayerItems *pb.GetPlayerItemsResponse
		if rGetPlayerItems, err = services.NnSDK.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
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

		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Avanzo di stage
		c.CurrentState.Stage = 2

	// In questo stage riepilogo le risorse necessarie e
	// chiedo al conferma al player se continuare il crafting dell'item
	case 2:
		// Inserisco nel payload la recipelist per avere accesso più facile ad essa
		helpers.UnmarshalPayload(c.Payload.Item.Recipe.RecipeList, &c.Payload.Resources)

		// Genero string contenente le risorse richieste per il craft
		var itemsRecipeList string
		for resourceID, value := range c.Payload.Resources {
			var rGetResourceByID *pb.GetResourceByIDResponse
			if rGetResourceByID, err = services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: resourceID,
			}); err != nil {
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
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3

	// In questo stage mi aspetto che l'utente abbia confermato e se così fosse
	// procedo con il rimuovere le risorse associate e notificargli l'attesa per il crafting
	case 3:
		var rLaboratoryStartCrafting *pb.LaboratoryStartCraftingResponse
		if rLaboratoryStartCrafting, err = services.NnSDK.LaboratoryStartCrafting(helpers.NewContext(1), &pb.LaboratoryStartCraftingRequest{
			PlayerID: c.Player.GetID(),
			ItemID:   c.Payload.Item.ID,
		}); err != nil {
			return fmt.Errorf("error start laboratory crafting: %s", err.Error())
		}

		// Converto time
		var finishAt time.Time
		finishAt, err = ptypes.Timestamp(rLaboratoryStartCrafting.GetCraftingEndTime())
		if err != nil {
			panic(err)
		}

		// Invio messaggio
		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.wait", finishAt.Format("15:04:05")),
		)
		msg.ParseMode = "markdown"
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorna stato
		c.CurrentState.Stage = 4
		c.ForceBackTo = true

	// In questo stage il player ha completato correttamente il crafting, quindi
	// proseguo con l'assegnarli l'item e concludo
	case 4:
		var rLaboratoryEndCrafting *pb.LaboratoryEndCraftingResponse
		if rLaboratoryEndCrafting, err = services.NnSDK.LaboratoryEndCrafting(helpers.NewContext(1), &pb.LaboratoryEndCraftingRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			return fmt.Errorf("error end laboratory crafting: %s", err.Error())
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(
				c.Player.Language.Slug,
				"ship.laboratory.craft_completed",
				helpers.Trans(c.Player.Language.Slug, "items."+rLaboratoryEndCrafting.GetItem().GetSlug()),
			),
		)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
