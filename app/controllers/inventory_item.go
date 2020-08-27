package controllers

import (
	"fmt"
	"strings"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// InventoryItemController
// Con questo controller il player avrà la possibilità di usare gli item
// da lui craftati e non. Quindi di beneficiare dei potenziamenti.
// ====================================
type InventoryItemController struct {
	BaseController
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
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller: "route.inventory.items",
		ControllerBack: ControllerBack{
			To:        &InventoryController{},
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
func (c *InventoryItemController) Validator() (hasErrors bool) {
	var err error
	switch c.PlayerData.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	// Verifico quale item ha scelto di usare e controllo se il player ha realmente
	// l'item indicato
	case 1:
		var rGetPlayerItems *pb.GetPlayerItemsResponse
		rGetPlayerItems, err = services.NnSDK.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return false
		}

		// Recupero nome item che il player vuole usare
		playerChoiche := strings.Split(c.Update.Message.Text, " (")[0]

		for _, item := range rGetPlayerItems.GetPlayerInventory() {
			if playerChoiche == helpers.Trans(c.Player.Language.Slug, "items."+item.Item.Slug) {
				c.Payload.Item = item.GetItem()
				return false
			}
		}

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
	switch c.PlayerData.CurrentState.Stage {

	// In questo stage recupero tutti gli item del player e li riporto sul tastierino
	case 0:
		// Recupero items del player
		var rGetPlayerItems *pb.GetPlayerItemsResponse
		rGetPlayerItems, err = services.NnSDK.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			panic(err)
		}

		// Ciclo items e li inserisco nella keyboarc
		var keyboardRowItems [][]tgbotapi.KeyboardButton
		for _, item := range rGetPlayerItems.GetPlayerInventory() {
			keyboardRowItem := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					fmt.Sprintf(
						"%s (%v)",
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
		msg := services.NewMessage(
			c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "inventory.items.what"),
		)

		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowItems,
		}

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Avanzo di stage
		c.PlayerData.CurrentState.Stage = 1

	// In questo stage chiedo conferma al player dell'item che itende usare
	case 1:

		var text string
		if c.PlayerData.PlayerStats.GetLifePoint()+c.Payload.Item.Value > 100 {
			text = fmt.Sprintf(
				"%s\n\n%s", // Domanda e descrizione
				helpers.Trans(c.Player.Language.Slug, "inventory.items.confirm_warning",
					helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
				),
				helpers.Trans(c.Player.Language.Slug, "items.description."+c.Payload.Item.Slug, c.Payload.Item.Value),
			)
		} else {
			text = fmt.Sprintf(
				"%s\n\n%s", // Domanda e descrizione
				helpers.Trans(c.Player.Language.Slug, "inventory.items.confirm",
					helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
				),
				helpers.Trans(c.Player.Language.Slug, "items.description."+c.Payload.Item.Slug, c.Payload.Item.Value),
			)
		}

		msg := services.NewMessage(c.Update.Message.Chat.ID, text)

		msg.ParseMode = "markdown"

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.PlayerData.CurrentState.Stage = 2

	// In questo stage se l'utente ha confermato continuo con con la richiesta
	case 2:
		// Richiamo il ws per usare l'item selezionato
		_, err := services.NnSDK.UseItem(helpers.NewContext(1), &pb.UseItemRequest{
			PlayerID: c.Player.ID,
			ItemID:   c.Payload.Item.ID,
		})
		if err != nil {
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "inventory.items.completed",
				helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
			),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.PlayerData.CurrentState.Completed = true
	}

	return
}
