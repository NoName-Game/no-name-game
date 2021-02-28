package controllers

import (
	"fmt"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerInventoryPackController
// ====================================
// Con questo controller il player avrà la possibilità di aprire i pacchetti
// Quindi di beneficiare dei drops.
// ====================================

type PlayerInventoryPackController struct {
	Controller
	Payload struct {
		Item *pb.Item
	}
}

// ====================================
// Handle
// ====================================
func (c *PlayerInventoryPackController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.inventory.packs",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerInventoryController{},
				FromStage: 1,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
				2: {"route.breaker.menu", "route.breaker.back"},
			},
		},
	}) {
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
func (c *PlayerInventoryPackController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico quale item ha scelto di usare e controllo se il player possiede realmente l'item indicato
	// ##################################################################################################
	case 1:
		var err error
		var rGetPlayerPacks *pb.GetPlayerPacksResponse
		if rGetPlayerPacks, err = config.App.Server.Connection.GetPlayerPacks(helpers.NewContext(1), &pb.GetPlayerPacksRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero nome item che il player vuole usare
		var itemChoosed string
		itemSplit := strings.Split(c.Update.Message.Text, " (")
		if len(itemSplit)-1 > 0 {
			itemSplit = strings.Split(itemSplit[0], " - ")
			if len(itemSplit)-1 > 0 {
				itemChoosed = itemSplit[1]
			}
		}

		for _, item := range rGetPlayerPacks.GetPlayerInventory() {
			if itemChoosed == helpers.Trans(c.Player.Language.Slug, "items."+item.Item.Slug) {
				c.Payload.Item = item.GetItem()
				return false
			}
		}

		return true

	// ##################################################################################################
	// Verifico la conferma dell'uso
	// ##################################################################################################
	case 2:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *PlayerInventoryPackController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// In questo stage recupero tutti gli item del player e li riporto sul tastierino
	case 0:
		// Recupero items del player
		var rGetPlayerPacks *pb.GetPlayerPacksResponse
		if rGetPlayerPacks, err = config.App.Server.Connection.GetPlayerPacks(helpers.NewContext(1), &pb.GetPlayerPacksRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Ciclo items e li inserisco nella keyboard
		var keyboardRowItems [][]tgbotapi.KeyboardButton

		// Sorting inventario
		inv := helpers.SortItemByCategory(rGetPlayerPacks.GetPlayerInventory())

		for _, item := range inv {
			keyboardRowItem := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					fmt.Sprintf(
						"%s - %s (%v)",
						helpers.Trans(c.Player.Language.Slug, "items."+item.Item.ItemCategory.Slug),
						helpers.Trans(c.Player.Language.Slug, "items."+item.Item.Slug),
						item.Quantity,
					),
				),
			)

			if item.Quantity > 0 {
				keyboardRowItems = append(keyboardRowItems, keyboardRowItem)
			}
		}

		keyboardRowItems = append(keyboardRowItems, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		// Invio messagio con recap e con selettore categoria
		msg := helpers.NewMessage(
			c.ChatID,
			helpers.Trans(c.Player.Language.Slug, "inventory.packs.what"),
		)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowItems,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1

	// In questo stage chiedo conferma al player dell'item che itende usare
	case 1:
		var text string

		text = fmt.Sprintf(
			"%s\n%s", // Domanda e descrizione
			helpers.Trans(c.Player.Language.Slug, "inventory.items.confirm",
				helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
			),
			helpers.Trans(c.Player.Language.Slug, "items.description."+c.Payload.Item.Slug),
		)

		msg := helpers.NewMessage(c.ChatID, text)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2

	// In questo stage se l'utente ha confermato continuo con con la richiesta
	case 2:
		// Richiamo il ws per usare l'item selezionato
		var drops *pb.OpenPackResponse
		if drops, err = config.App.Server.Connection.OpenPack(helpers.NewContext(1), &pb.OpenPackRequest{
			PlayerID: c.Player.ID,
			ItemID:   c.Payload.Item.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}
		var dropResults string
		// Creo messaggio di recap drop
		for _, resource := range drops.GetResources() {
			dropResults += fmt.Sprintf(
				"- %s %s (<b>%s</b>) %s\n",
				helpers.GetResourceCategoryIcons(resource.GetResourceCategoryID()),
				resource.GetName(),
				strings.ToUpper(resource.GetRarity().GetSlug()),
				helpers.GetResourceBaseIcons(resource.GetBase()))
		}
		if len(drops.GetItems()) > 0 {
			dropResults += "\n"
			for _, item := range drops.Items {
				dropResults += fmt.Sprintf("- <b>%s</b>\n", helpers.Trans(c.Player.Language.Slug, "items."+item.GetSlug()))
			}
		}
		// Countdown 3-2-1 Drop

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID,
			helpers.Trans(c.Player.Language.Slug, "inventory.packs.completed", dropResults))
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true

		// ###################
		// TUTORIAL - Solo il player si trova dentro il tutorial forzo di tornarare al menu
		// ###################
		if c.InTutorial() {
			c.Configurations.ControllerBack.To = &MenuController{}
		}
	}

	return
}
