package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerInventoryItemController
// ====================================
// Con questo controller il player avrà la possibilità di usare gli item
// da lui craftati e non. Quindi di beneficiare dei potenziamenti.
// ====================================

type PlayerInventoryItemController struct {
	Controller
	Payload struct {
		Item     *pb.Item
		Quantity int32
	}
}

func (c *PlayerInventoryItemController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.inventory.items",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerInventoryController{},
				FromStage: 1,
			},
			BreakerPerStage: map[int32][]string{
				1: {"route.breaker.menu"},
				2: {"route.breaker.back"},
				3: {"route.breaker.back"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *PlayerInventoryItemController) Handle(player *pb.Player, update tgbotapi.Update) {
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
func (c *PlayerInventoryItemController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico quale item ha scelto di usare e controllo se il player possiede realmente l'item indicato
	// ##################################################################################################
	case 1:
		var err error
		var rGetPlayerItems *pb.GetPlayerItemsResponse
		if rGetPlayerItems, err = config.App.Server.Connection.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
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

		for _, item := range rGetPlayerItems.GetPlayerInventory() {
			if itemChoosed == helpers.Trans(c.Player.Language.Slug, "items."+item.Item.Slug) {
				c.Payload.Item = item.GetItem()
				return false
			}
		}

		return true
	case 2:
		// Quantità
		quantity, err := strconv.Atoi(c.Update.Message.Text)
		if err != nil || quantity < 1 || quantity > 50 {
			return true
		}

		c.Payload.Quantity = int32(quantity)
	// ##################################################################################################
	// Verifico la conferma dell'uso
	// ##################################################################################################
	case 3:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *PlayerInventoryItemController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// In questo stage recupero tutti gli item del player e li riporto sul tastierino
	case 0:
		// Recupero items del player
		var rGetPlayerItems *pb.GetPlayerItemsResponse
		if rGetPlayerItems, err = config.App.Server.Connection.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Ciclo items e li inserisco nella keyboard
		var keyboardRowItems [][]tgbotapi.KeyboardButton

		// Sorting inventario
		inv := helpers.SortItemByCategory(rGetPlayerItems.GetPlayerInventory())

		for _, item := range inv {
			// Rimuovo amuleti dalla visualizzazione
			// Nel caso diventassero pìu oggetti creare un metodo dedicato
			if item.Item.ID == 7 || item.Item.ID > 9 {
				continue
			}

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
			helpers.Trans(c.Player.Language.Slug, "inventory.items.what"),
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
	case 1:
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "inventory.items.many"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("2"),
				tgbotapi.NewKeyboardButton("5"),
			),
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back"))),
		)
		if _, err := helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 2

	// In questo stage chiedo conferma al player dell'item che itende usare
	case 2:
		var text string

		text = fmt.Sprintf(
			"%s\n%s", // Domanda e descrizione
			helpers.Trans(c.Player.Language.Slug, "inventory.items.confirm",
				helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
			),
			helpers.Trans(c.Player.Language.Slug, "items.description."+c.Payload.Item.Slug, c.Payload.Item.Value),
		)

		// Verifica eccedenza
		if int32(c.Player.GetLifePoint())+c.Payload.Item.Value > 100 {
			text += helpers.Trans(c.Player.Language.Slug, "inventory.items.confirm_warning")
		}

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
		c.CurrentState.Stage = 3

	// In questo stage se l'utente ha confermato continuo con con la richiesta
	case 3:
		// Richiamo il ws per usare l'item selezionato
		if _, err := config.App.Server.Connection.UseItem(helpers.NewContext(1), &pb.UseItemRequest{
			PlayerID: c.Player.ID,
			ItemID:   c.Payload.Item.ID,
			Quantity: uint32(c.Payload.Quantity),
		}); err != nil {
			var msg tgbotapi.MessageConfig
			if strings.Contains(err.Error(), "item quantity less than zero") {
				// Invia messaggio
				msg = helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "inventory.items.none"))
			} else {
				c.Logger.Panic(err)
			}
			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}
			c.CurrentState.Completed = true
			return
		}

		var text string
		text = helpers.Trans(c.Player.Language.Slug, "inventory.items.completed",
			helpers.Trans(c.Player.Language.Slug, "items."+c.Payload.Item.Slug),
		)
		if c.Payload.Item.GetItemCategory().GetSlug() == "ship_support" {
			// Riporto le informazioni della nave post item
			var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
			if rGetPlayerShipEquipped, err = config.App.Server.Connection.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{PlayerID: c.Player.ID}); err != nil {
				c.Logger.Panic(err)
			}
			text += fmt.Sprintf("\n\n🚀 <b>%s</b> (<b>%s</b>) - <b>%s</b>\n🔧 <b>%v%%</b> (%s)\n⛽ <b>%v%%</b> (%s)",
				rGetPlayerShipEquipped.GetShip().GetName(), strings.ToUpper(rGetPlayerShipEquipped.GetShip().GetRarity().GetSlug()),
				rGetPlayerShipEquipped.GetShip().GetShipCategory().GetName(),
				rGetPlayerShipEquipped.GetShip().GetIntegrity(), helpers.Trans(c.Player.Language.Slug, "integrity"),
				rGetPlayerShipEquipped.GetShip().GetTank(), helpers.Trans(c.Player.Language.Slug, "fuel"),
			)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID, text)
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
