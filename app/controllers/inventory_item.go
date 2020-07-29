package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

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
func (c *InventoryItemController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.inventory.items",
		c.Payload,
		[]string{},
		player,
		update,
	) {
		return
	}

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(1, &InventoryController{}) {
			return
		}
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
	if hasError {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
		validatorMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := services.NnSDK.UpdatePlayerState(ctx, &pb.UpdatePlayerStateRequest{
		PlayerState: c.State,
	})
	if err != nil {
		panic(err)
	}
	c.State = response.GetPlayerState()

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *InventoryItemController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		),
	)

	switch c.State.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	// Verifico quale item ha scelto di usare e controllo se il player ha realmente
	// l'item indicato
	case 1:
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		response, err := services.NnSDK.GetPlayerItems(ctx, &pb.GetPlayerItemsRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return false, err
		}

		var playerInventoryItems []*pb.PlayerInventory
		playerInventoryItems = response.GetPlayerInventory()

		// Recupero nome item che il player vuole usare
		playerChoiche := strings.Split(c.Update.Message.Text, " (")[0]

		for _, item := range playerInventoryItems {
			if playerChoiche == helpers.Trans(c.Player.Language.Slug, "items."+item.Item.Slug) {
				c.Payload.Item = item.GetItem()
				return false, err
			}
		}

		return true, err

	// Verifico la conferma dell'uso
	case 2:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "confirm") {
			return false, err
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true, err
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *InventoryItemController) Stage() (err error) {
	switch c.State.Stage {

	// In questo stage recupero tutti gli item del player e li riporto sul tastierino
	case 0:
		// Recupero items del player
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		response, err := services.NnSDK.GetPlayerItems(ctx, &pb.GetPlayerItemsRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			panic(err)
		}

		var playerInventoryItems []*pb.PlayerInventory
		playerInventoryItems = response.GetPlayerInventory()

		// Ciclo items e li inserisco nella keyboarc
		var keyboardRowItems [][]tgbotapi.KeyboardButton
		for _, item := range playerInventoryItems {
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
		c.State.Stage = 1

	// In questo stage chiedo conferma al player dell'item che itende usare
	case 1:
		var text string
		if c.Player.GetStats().GetLifePoint()+uint32(c.Payload.Item.Value) > 100 {
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
		c.State.Stage = 2

	// In questo stage se l'utente ha confermato continuo con con la richiesta
	case 2:
		// Richiamo il ws per usare l'item selezionato
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := services.NnSDK.UseItem(ctx, &pb.UseItemRequest{
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
		c.State.Completed = true
	}

	return
}
