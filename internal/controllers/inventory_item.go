package controllers

import (
	"fmt"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/config"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// InventoryItemController
// ====================================
// Con questo controller il player avrà la possibilità di usare gli item
// da lui craftati e non. Quindi di beneficiare dei potenziamenti.
// ====================================

type InventoryItemController struct {
	Controller
	Payload struct {
		Item *pb.Item
	}
}

// ====================================
// Handle
// ====================================
func (c *InventoryItemController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.inventory.items",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &InventoryController{},
				FromStage: 1,
			},
		},
	}) {
		return
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
func (c *InventoryItemController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	// Verifico quale item ha scelto di usare e controllo se il player possiede realmente l'item indicato
	case 1:
		var rGetPlayerItems *pb.GetPlayerItemsResponse
		if rGetPlayerItems, err = config.App.Server.Connection.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			return false
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

		for _, item := range rGetPlayerItems.GetPlayerInventory() {
			if itemChoosed == helpers.Trans(c.Player.Language.Slug, "items."+item.Item.Slug) {
				c.Payload.Item = item.GetItem()
				return false
			}
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "inventory.items.not_found")
		return true

	// Verifico la conferma dell'uso
	case 2:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "confirm") {
			return false
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *InventoryItemController) Stage() (err error) {
	switch c.CurrentState.Stage {

	// In questo stage recupero tutti gli item del player e li riporto sul tastierino
	case 0:
		// Recupero items del player
		var rGetPlayerItems *pb.GetPlayerItemsResponse
		if rGetPlayerItems, err = config.App.Server.Connection.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			return err
		}

		// Ciclo items e li inserisco nella keyboard
		var keyboardRowItems [][]tgbotapi.KeyboardButton
		for _, item := range rGetPlayerItems.GetPlayerInventory() {
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

			keyboardRowItems = append(keyboardRowItems, keyboardRowItem)
		}

		keyboardRowItems = append(keyboardRowItems, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messagio con recap e con selettore categoria
		msg := helpers.NewMessage(
			c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "inventory.items.what"),
		)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowItems,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			return err
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
			helpers.Trans(c.Player.Language.Slug, "items.description."+c.Payload.Item.Slug, c.Payload.Item.Value),
		)

		// Verifica eccedenza
		if c.Data.PlayerStats.GetLifePoint()+c.Payload.Item.Value > 100 {
			text += helpers.Trans(c.Player.Language.Slug, "inventory.items.confirm_warning")
		}

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, text)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2

	// In questo stage se l'utente ha confermato continuo con con la richiesta
	case 2:
		// Richiamo il ws per usare l'item selezionato
		if _, err := config.App.Server.Connection.UseItem(helpers.NewContext(1), &pb.UseItemRequest{
			PlayerID: c.Player.ID,
			ItemID:   c.Payload.Item.ID,
		}); err != nil {
			return err
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "inventory.items.completed",
				helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
			),
		)
		msg.ParseMode = "markdown"
		_, err = helpers.SendMessage(msg)
		if err != nil {
			return err
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
