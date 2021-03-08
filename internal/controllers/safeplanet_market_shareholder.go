package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
)

// ====================================
// SafePlanetMarketDealerController
// ====================================
type SafePlanetMarketShareHolderController struct {
	Payload struct {
		ActionID uint32
		ResourceID uint32
		SystemID uint32
		Quantity int32
		Action   string // BUY SELL
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMarketShareHolderController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.market.shareholder",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetMarketController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
				2: {"route.breaker.back","route.breaker.menu","route.breaker.clears"},
				3: {"route.breaker.back","route.breaker.menu","route.breaker.clears"},
				4: {"route.breaker.menu","route.breaker.clears"},
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
func (c *SafePlanetMarketShareHolderController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	case 1:
		// Controllo che il nome della risorsa sia presente nel mercato
		var rGetSystemActions *pb.GetActionsBySystemIDResponse
		if rGetSystemActions, err = config.App.Server.Connection.GetActionsBySystemID(helpers.NewContext(1), &pb.GetActionsBySystemIDRequest{
			SystemID: c.Payload.SystemID,
		}); err != nil {
			c.Logger.Panic(err)
		}
		for _, action := range rGetSystemActions.GetActions() {
			if action.GetResource().GetName() == c.Update.Message.Text {
				c.Payload.ActionID = action.ID
				c.Payload.ResourceID = action.ResourceID
				return false
			}
		}
		// Risorsa non trovata
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.error_no_action")
		c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.menu"),
				),
			),
		)
		return true
	case 2:
		// Controllo la scelta fatta
		switch c.Update.Message.Text {
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.buy"):
			c.Payload.Action = "buy"
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.sell"):
			c.Payload.Action = "sell"
		default:
			return true
		}
	case 3:
		quantity, err := strconv.Atoi(c.Update.Message.Text)
		if err != nil || quantity < 1 || quantity > 50 {
			return true
		}

		c.Payload.Quantity = int32(quantity)

		// Controllo anche che ci siano abbastanza risorse per quell'azione
		var rGetActionByID *pb.GetActionByIDResponse
		if rGetActionByID, err = config.App.Server.Connection.GetActionByID(helpers.NewContext(1), &pb.GetActionByIDRequest{ID: c.Payload.ActionID}); err != nil {
			c.Logger.Panic(err)
		}
		if rGetActionByID.GetQuantity() < c.Payload.Quantity  && c.Payload.Action == "buy" {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.error_no_resources_left")
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.menu"),
					),
				),
			)
			return true
		}
	case 4:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetMarketShareHolderController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		// Recupero posizione player corrente
		var playerPosition *pb.Planet
		if playerPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
			c.Logger.Panic(err)
		}
		// Mostro tutte le risorse disponibili nel mercato di quel sistema
		var rGetSystemActions *pb.GetActionsBySystemIDResponse
		if rGetSystemActions, err = config.App.Server.Connection.GetActionsBySystemID(helpers.NewContext(1), &pb.GetActionsBySystemIDRequest{
			SystemID: playerPosition.GetPlanetSystemID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		c.Payload.SystemID = playerPosition.GetPlanetSystemID()

		var keyboard [][]tgbotapi.KeyboardButton

		var text string
		text += helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.intro")
		for _, action := range rGetSystemActions.GetActions() {
			// Per ogni azione recupero tutte le informazioni, TODO da controllare l'efficacia di sto giro di request.
			var rGetActionByID *pb.GetActionByIDResponse
			if rGetActionByID, err = config.App.Server.Connection.GetActionByID(helpers.NewContext(1), &pb.GetActionByIDRequest{ID: action.GetID()}); err != nil {
				c.Logger.Panic(err)
			}
			var avaible string
			if rGetActionByID.GetQuantity() > 0 {
				avaible = "\U0001F7E2"
			} else {
				avaible = "🔴"
			}
			var trend string
			if rGetActionByID.GetCurrentValue() > action.GetStartingPrice() {
				trend = "🔺"
			} else {
				trend = "🔻"
			}
			text += fmt.Sprintf("%s <code>%d x %s (%s) ~%d</code> %s\n", avaible,rGetActionByID.GetQuantity(), action.Resource.GetName(), action.GetResource().GetRarity().GetSlug(), rGetActionByID.GetCurrentValue(), trend)
			keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(action.GetResource().GetName())))
		}

		keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		msg := helpers.NewMessage(c.Player.ChatID, text)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboard,
			ResizeKeyboard: true,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1
	case 1:
		var rGetPlayerResource *pb.GetPlayerResourceByIDResponse
		if rGetPlayerResource, err = config.App.Server.Connection.GetPlayerResourceByID(helpers.NewContext(1), &pb.GetPlayerResourceByIDRequest{
			PlayerID:   c.Player.ID,
			ResourceID: c.Payload.ResourceID,
		}); err != nil {
			c.Logger.Panic(err)
		}
		// Ora chiedo cosa vuole effettuare (BUY/SELL)
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.what", c.Update.Message.Text, rGetPlayerResource.GetPlayerInventory().GetQuantity()))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.buy")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.sell")),
			),
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back"))),
		)
		if _, err := helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 2
	case 2:
		// Ora chiedo la quantità di risorse con cui vuole operare
		// Ora chiedo cosa vuole effettuare (BUY/SELL)
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.quantity", strings.ToLower(helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder."+c.Payload.Action))))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("5"),
				tgbotapi.NewKeyboardButton("10"),
				tgbotapi.NewKeyboardButton("20"),
			),
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back"))),
		)
		if _, err := helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 3
	case 3:
		var t pb.GetBidInfoRequest_BidTypeEnum
		switch c.Payload.Action {
		case "buy":
			t = pb.GetBidInfoRequest_BUY
		case "sell":
			t = pb.GetBidInfoRequest_SELL
		}
		// Recupero info sulla transazione che sta per effettuare
		var rGetBidInfo *pb.GetBidInfoResponse
		if rGetBidInfo, err = config.App.Server.Connection.GetBidInfo(helpers.NewContext(1), &pb.GetBidInfoRequest{
			ActionID: c.Payload.ActionID,
			Quantity: c.Payload.Quantity,
			Type: t,
		}); err != nil {
			c.Logger.Panic(err)
		}
		// Chiedo conferma e stampo alcuni dati relativi alla transazione
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.confirm_"+c.Payload.Action, rGetBidInfo.GetValue()))
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

		c.CurrentState.Stage = 4
	case 4:
		var t pb.PlaceBidRequest_BidTypeEnum
		switch c.Payload.Action {
		case "buy":
			t = pb.PlaceBidRequest_BUY
		case "sell":
			t = pb.PlaceBidRequest_SELL
		}
		// E' tempo di effettuare la BID
		var rPlaceBid *pb.PlaceBidResponse
		if rPlaceBid, err = config.App.Server.Connection.PlaceBid(helpers.NewContext(1), &pb.PlaceBidRequest{
			ActionID: c.Payload.ActionID,
			PlayerID: c.Player.ID,
			Quantity: uint32(c.Payload.Quantity),
			Type:     t,
		}); err != nil {
			var msg tgbotapi.MessageConfig
			if strings.Contains(err.Error(), "player dont have enough money") {
				msg = helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.error_no_money"))
			} else if strings.Contains(err.Error(), "item quantity less than zero") || strings.Contains(err.Error(), "no resource") {
				msg = helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safaplanet.shareholder.error_no_resources"))
			}
			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}
			c.CurrentState.Completed = true
			return
		}
		// Se tutto è andato a buon fine crafto il messaggio finale
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.shareholder.complete."+c.Payload.Action, c.Payload.Quantity, rPlaceBid.GetResource().GetName(), rPlaceBid.GetValue()))
		msg.ParseMode = tgbotapi.ModeHTML

		if _, err := helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}
		c.CurrentState.Completed = true
	}

	return
}
