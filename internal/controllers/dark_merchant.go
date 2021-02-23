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
// DarkMerchantController
// ====================================
type DarkMerchantController struct {
	Payload struct {
		ResourceID uint32
		Price      int64
		Quantity   int64
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *DarkMerchantController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.darkmerchant",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			PlanetType: []string{"darkMerchant"},
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
func (c *DarkMerchantController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico quale item vuole comprare il player
	// ##################################################################################################
	case 1:
		var err error
		var rGetDarkMerchant *pb.GetDarkMerchantResponse
		if rGetDarkMerchant, err = config.App.Server.Connection.GetDarkMerchant(helpers.NewContext(1), &pb.GetDarkMerchantRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		resourceName := strings.Split(c.Update.Message.Text, " ")[1]
		for _, resource := range rGetDarkMerchant.GetResources() {
			if resourceName == resource.GetResource().GetName() {
				c.Payload.ResourceID = resource.GetResource().GetID()

				// Salvo anche prezzo della risorsa scelta
				c.Payload.Price = resource.GetPrice()

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

		c.Payload.Quantity = int64(quantity)

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
func (c *DarkMerchantController) Stage() {
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

		var rGetDarkMerchant *pb.GetDarkMerchantResponse
		if rGetDarkMerchant, err = config.App.Server.Connection.GetDarkMerchant(helpers.NewContext(1), &pb.GetDarkMerchantRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		var itemsKeyboard [][]tgbotapi.KeyboardButton
		for _, resource := range rGetDarkMerchant.GetResources() {
			itemsKeyboard = append(itemsKeyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					fmt.Sprintf("%s %s (%s) %s - 💰%v",
						helpers.GetResourceCategoryIcons(resource.GetResource().GetResourceCategoryID()),
						resource.GetResource().GetName(), // Nome risorsa
						resource.GetResource().GetRarity().GetSlug(),
						helpers.GetResourceBaseIcons(resource.GetResource().GetBase()),
						resource.Price, // Costo
					),
				),
			))
		}

		// Aggiungo torna al menu
		itemsKeyboard = append(itemsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "darkmerchant.intro", rGetPlayerEconomy.GetValue()))
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
		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "darkmerchant.quantity"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("2"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("5"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("10"),
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

	// ##################################################################################################
	// Faccio Recap e chiedo conferma
	// ##################################################################################################
	case 2:
		// Recupero dettagli item scelto
		var rGetResourceByID *pb.GetResourceByIDResponse
		if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
			ID: c.Payload.ResourceID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "darkmerchant.confirm",
			c.Payload.Quantity,
			rGetResourceByID.GetResource().GetName(), // Nome Risorsa
			c.Payload.Quantity*c.Payload.Price,       // Costo
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
		_, err = config.App.Server.Connection.DarkMerchantBuyResource(helpers.NewContext(1), &pb.DarkMerchantBuyResourceRequest{
			PlayerID:   c.Player.GetID(),
			ResourceID: c.Payload.ResourceID,
			Price:      c.Payload.Price,
			Quantity:   c.Payload.Quantity,
		})

		if err != nil && strings.Contains(err.Error(), "player dont have enough money") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di amuleti
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "darkmerchant.not_enough_money"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		} else if err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "darkmerchant.completed"))
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
