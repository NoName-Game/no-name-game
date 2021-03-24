package controllers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipLaboratoryController
// Ogni player ha la possibilità di craftare al player degli item
// che possono essere usati in diversi modo, es. per recuperare vita
// o per ripristinare determinate cose
// ====================================
type ShipLaboratoryController struct {
	Controller
	Payload struct {
		CategoryID uint32
		ItemID     uint32
		// Resources  map[uint32]int32 // Materiali necessari
	}
}

func (c *ShipLaboratoryController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship.laboratory",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBlocked: []string{"exploration", "hunting"},
			ControllerBack: ControllerBack{
				To:        &ShipController{},
				FromStage: 1,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
				2: {"route.breaker.back"},
				3: {"route.breaker.back"},
				4: {"route.breaker.continue"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipLaboratoryController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Validate
	if c.Validator() {
		c.Validate()
		return
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *ShipLaboratoryController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// In questo stage verifico se mi è stata passata una categoria che esiste realmente
	// ##################################################################################################
	case 1:
		var err error

		// Recupero tutte le categorie degli items e ciclo per trovare quella voluta del player
		var rGetAllItemCategories *pb.GetAllItemCategoriesResponse
		if rGetAllItemCategories, err = config.App.Server.Connection.GetAllItemCategories(helpers.NewContext(1), &pb.GetAllItemCategoriesRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		for _, category := range rGetAllItemCategories.GetItemCategories() {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ship.laboratory.categories.%s", category.Slug)) {
				c.Payload.CategoryID = category.GetID()
				return false
			}
		}

		return true

	// ##################################################################################################
	// Recupero tutte gli items e ciclo per trovare quello voluta del player
	// ##################################################################################################
	case 2:
		var err error
		var rGetAllItems *pb.GetAllItemsResponse
		if rGetAllItems, err = config.App.Server.Connection.GetAllItems(helpers.NewContext(1), &pb.GetAllItemsRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero nome item che il player vuole craftare
		playerChoiche := strings.Split(c.Update.Message.Text, " (")[0]
		var itemExists bool
		for _, item := range rGetAllItems.GetItems() {
			if playerChoiche == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", item.Slug)) {
				itemExists = true
				c.Payload.ItemID = item.GetID()
			}
		}

		// Se l'item non esiste ritorno errore
		if !itemExists {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.laboratory.item_does_not_exist")
			return true
		}

	// ##################################################################################################
	// In questo stage è necessario che venga validato se il player ha tutti i
	// materiali necessario al crafting dell'item da lui scelto
	// ##################################################################################################
	case 3:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "yep") {
			var err error
			var rLaboratoryCheckHaveResourceForCrafting *pb.LaboratoryCheckHaveResourceForCraftingResponse
			if rLaboratoryCheckHaveResourceForCrafting, err = config.App.Server.Connection.LaboratoryCheckHaveResourceForCrafting(helpers.NewContext(1), &pb.LaboratoryCheckHaveResourceForCraftingRequest{
				PlayerID: c.Player.ID,
				ItemID:   c.Payload.ItemID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			if !rLaboratoryCheckHaveResourceForCrafting.GetHaveResources() {
				c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.laboratory.no_resource_to_craft")
				c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
						),
					),
				)

				return true
			}

			return false
		}

		return true

	// ##################################################################################################
	// In questo stage verificho che l'utente abbia effettivamente aspettato
	// il tempo di attesa necessario al craft
	// ##################################################################################################
	case 4:
		var err error
		var rLaboratoryCheckCrafting *pb.LaboratoryCheckCraftingResponse
		if rLaboratoryCheckCrafting, err = config.App.Server.Connection.LaboratoryCheckCrafting(helpers.NewContext(1), &pb.LaboratoryCheckCraftingRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Il crafter sta già portando a terminre un lavoro per questo player
		if !rLaboratoryCheckCrafting.GetFinishCrafting() {
			var finishAt time.Time
			if finishAt, err = helpers.GetEndTime(rLaboratoryCheckCrafting.GetCraftingEndTime(), c.Player); err != nil {
				c.Logger.Panic(err)
			}

			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"ship.laboratory.wait_validator",
				finishAt.Format("15:04:05"),
			)
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.continue"),
					),
				),
			)

			return true
		}
	}

	return false
}

// ====================================
// Stage  0 What -> 1 - Check Resources -> 2 - Confirm -> 3 - Craft
// ====================================
func (c *ShipLaboratoryController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Lista di categorie item craftabili
	// ##################################################################################################
	case 0:
		var rGetAllItemCategories *pb.GetAllItemCategoriesResponse
		if rGetAllItemCategories, err = config.App.Server.Connection.GetAllItemCategories(helpers.NewContext(1), &pb.GetAllItemCategoriesRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		// Creo messaggio
		msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s\n\n%s",
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.type"),
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.info"),
		))

		var keyboardRow [][]tgbotapi.KeyboardButton
		for _, category := range rGetAllItemCategories.GetItemCategories() {
			// Tolgo momentaneamente Altro dalla pagina di crafting, rimuovere quando ci saranno item craftabili.
			if category.Slug == "stuff" || category.Slug == "pack" {
				continue
			}
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
				helpers.Trans(c.Player.Language.Slug, "route.breaker.menu"),
			),
		))

		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRow,
			ResizeKeyboard: true,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1

	// ##################################################################################################
	// Recuerpero Lista oggetti craftabili in base alla categoria scelta dal player
	// ##################################################################################################
	case 1:
		var rGetCraftableItemsByCategoryID *pb.GetCraftableItemsByCategoryIDResponse
		if rGetCraftableItemsByCategoryID, err = config.App.Server.Connection.GetCraftableItemsByCategoryID(helpers.NewContext(1), &pb.GetCraftableItemsByCategoryIDRequest{
			CategoryID: c.Payload.CategoryID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Creo messaggio
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "ship.laboratory.what"))

		var keyboardRow [][]tgbotapi.KeyboardButton
		for _, item := range rGetCraftableItemsByCategoryID.GetItems() {
			// Recupero quantità del player per quest'item
			var rGetPlayerItemByID *pb.GetPlayerItemByIDResponse
			if rGetPlayerItemByID, err = config.App.Server.Connection.GetPlayerItemByID(helpers.NewContext(1), &pb.GetPlayerItemByIDRequest{
				PlayerID: c.Player.ID,
				ItemID:   item.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			row := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					fmt.Sprintf(
						"%s (%v)",
						helpers.Trans(c.Player.Language.Slug, "items."+item.Slug),
						rGetPlayerItemByID.GetPlayerInventory().GetQuantity(),
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

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stage
		c.CurrentState.Stage = 2

	// ##################################################################################################
	// Riepilogo risorse necessarie e chiedo conferma
	// ##################################################################################################
	case 2:
		// Recupero ricetta item scelta
		var rGetItemByID *pb.GetItemByIDResponse
		if rGetItemByID, err = config.App.Server.Connection.GetItemByID(helpers.NewContext(1), &pb.GetItemByIDRequest{
			ItemID: c.Payload.ItemID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero dettaglio ricetta
		var recipe map[uint32]int32
		if err = json.Unmarshal([]byte(rGetItemByID.GetItem().GetItemRecipe().GetRecipeList()), &recipe); err != nil {
			c.Logger.Panic(err)
		}

		// Genero string contenente le risorse richieste per il craft
		var itemsRecipeList string
		for resourceID, value := range recipe {
			var rGetResourceByID *pb.GetResourceByIDResponse
			if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: resourceID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Verifico se il player possiede il quantitativo e la risorsa richiesta
			var rGetPlayerResourceByID *pb.GetPlayerResourceByIDResponse
			if rGetPlayerResourceByID, err = config.App.Server.Connection.GetPlayerResourceByID(helpers.NewContext(1), &pb.GetPlayerResourceByIDRequest{
				PlayerID:   c.Player.ID,
				ResourceID: rGetResourceByID.GetResource().GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Risultato verifica quantità
			haveQuantity := "❌"
			if rGetPlayerResourceByID.GetPlayerInventory().GetQuantity() >= value {
				haveQuantity = "✅"
			}

			// Se la risorsa non è stata ancora scoperta mostro ???
			resourceName := rGetResourceByID.GetResource().GetName()
			if !rGetResourceByID.GetResource().GetEnabled() {
				resourceName = "<b>???</b>"
			}

			itemsRecipeList += fmt.Sprintf("%s %s%s (%s) %s - %v/<b>%v</b>\n",
				haveQuantity,
				helpers.GetResourceCategoryIcons(rGetResourceByID.GetResource().GetResourceCategoryID()), resourceName,
				rGetResourceByID.GetResource().GetRarity().GetSlug(),
				helpers.GetResourceBaseIcons(rGetResourceByID.GetResource().GetBase()),
				rGetPlayerResourceByID.GetPlayerInventory().GetQuantity(), value,
			)
		}

		msg := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.you_need",
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", rGetItemByID.GetItem().GetSlug())),
				itemsRecipeList,
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.description.%s", rGetItemByID.GetItem().GetSlug()), rGetItemByID.GetItem().GetValue()),
			),
		)

		msg.ParseMode = tgbotapi.ModeHTML
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
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3

	// ##################################################################################################
	// Avvio crafting
	// ##################################################################################################
	case 3:
		var rLaboratoryStartCrafting *pb.LaboratoryStartCraftingResponse
		if rLaboratoryStartCrafting, err = config.App.Server.Connection.LaboratoryStartCrafting(helpers.NewContext(1), &pb.LaboratoryStartCraftingRequest{
			PlayerID: c.Player.GetID(),
			ItemID:   c.Payload.ItemID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Converto time
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rLaboratoryStartCrafting.GetCraftingEndTime(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "ship.laboratory.wait", finishAt.Format("15:04:05")),
		)

		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorna stato
		c.CurrentState.Stage = 4
		c.ForceBackTo = true

	// ##################################################################################################
	// Concludo Crafting
	// ##################################################################################################
	case 4:
		var rLaboratoryEndCrafting *pb.LaboratoryEndCraftingResponse
		if rLaboratoryEndCrafting, err = config.App.Server.Connection.LaboratoryEndCrafting(helpers.NewContext(1), &pb.LaboratoryEndCraftingRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"ship.laboratory.craft_completed",
				helpers.Trans(c.Player.Language.Slug, "items."+rLaboratoryEndCrafting.GetItem().GetSlug()),
			),
		)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
