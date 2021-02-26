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
// SafePlanetDealerController
// ====================================
type SafePlanetDealerController struct {
	Payload struct {
		ItemID   uint32
		Quantity int32
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetDealerController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.dealer",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
				2: {"route.breaker.back"},
				3: {"route.breaker.clears"},
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
func (c *SafePlanetDealerController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico quale item vuole comprare il player
	// ##################################################################################################
	case 1:
		var err error
		var rGetShoppableItems *pb.GetShoppableItemsResponse
		if rGetShoppableItems, err = config.App.Server.Connection.GetShoppableItems(helpers.NewContext(1), &pb.GetShoppableItemsRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		itemName := strings.Split(c.Update.Message.Text, " (")[0]

		for _, item := range rGetShoppableItems.GetItems() {
			if itemName == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", item.GetSlug())) {
				c.Payload.ItemID = item.GetID()
				return false
			}
		}

		return true

	// ##################################################################################################
	// Verifico se il quantitativo richiesto è valido
	// ##################################################################################################
	case 2:
		quantity, err := strconv.Atoi(c.Update.Message.Text)
		if err != nil || quantity < 1 {
			return true
		}

		c.Payload.Quantity = int32(quantity)

	// ##################################################################################################
	// Verifico conferma player
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
func (c *SafePlanetDealerController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Mostro item vendibili
	// ##################################################################################################
	case 0:
		// Recupero momente player, mi serve per mostrare budget
		var rGetPlayerEconomy *pb.GetPlayerEconomyResponse
		if rGetPlayerEconomy, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
			PlayerID:    c.Player.ID,
			EconomyType: pb.GetPlayerEconomyRequest_MONEY,
		}); err != nil {
			c.Logger.Panic(err)
		}

		var rGetShoppableItems *pb.GetShoppableItemsResponse
		if rGetShoppableItems, err = config.App.Server.Connection.GetShoppableItems(helpers.NewContext(1), &pb.GetShoppableItemsRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		var itemsKeyboard [][]tgbotapi.KeyboardButton
		for _, item := range rGetShoppableItems.GetItems() {
			itemsKeyboard = append(itemsKeyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					fmt.Sprintf("%s (💰%v)",
						helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", item.GetSlug())), // Nome item
						item.GetPrice(), // Costo
					),
				),
			))
		}

		// Aggiungo torna al menu
		itemsKeyboard = append(itemsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.dealer.intro", rGetPlayerEconomy.GetValue()))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       itemsKeyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1

	// ##################################################################################################
	// Chiedo quante ne vuole comprare
	// ##################################################################################################
	case 1:
		var keyboardRowQuantities [][]tgbotapi.KeyboardButton
		for i := 1; i <= 5; i++ {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(fmt.Sprintf("%d", i)),
			)
			keyboardRowQuantities = append(keyboardRowQuantities, keyboardRow)
		}

		// Aggiungo tasti back and clears
		keyboardRowQuantities = append(keyboardRowQuantities, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.dealer.quantity"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowQuantities,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2

	// ##################################################################################################
	// Faccio Recap e chiedo conferma
	// ##################################################################################################
	case 2:
		// Recupero dettagli item scelto
		var rGetItemByID *pb.GetItemByIDResponse
		if rGetItemByID, err = config.App.Server.Connection.GetItemByID(helpers.NewContext(1), &pb.GetItemByIDRequest{
			ItemID: c.Payload.ItemID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.dealer.confirm",
			c.Payload.Quantity,
			helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", rGetItemByID.GetItem().GetSlug())), // Nome item
			c.Payload.Quantity*rGetItemByID.GetItem().GetPrice(),                                             // Costo
		))

		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3

	// ##################################################################################################
	// Confermo acquisto
	// ##################################################################################################
	case 3:
		_, err = config.App.Server.Connection.BuyItem(helpers.NewContext(1), &pb.BuyItemRequest{
			PlayerID: c.Player.GetID(),
			ItemID:   c.Payload.ItemID,
			Quantity: c.Payload.Quantity,
		})

		if err != nil && strings.Contains(err.Error(), "player dont have enough money") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di amuleti
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.dealer.not_enough_money"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		} else if err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.dealer.completed"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
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
