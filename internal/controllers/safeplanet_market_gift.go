package controllers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"
)

// ====================================
// SafePlanetMarketGiftController
// ====================================
type SafePlanetMarketGiftController struct {
	Payload struct {
		ItemType       string
		ItemID         uint32
		Username       string
		ToPlayerChatID int64
		ToPlayerID     uint32
	}
	Controller
}

func (c *SafePlanetMarketGiftController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.market.gift",
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
				2: {"route.breaker.menu"},
				3: {"route.breaker.menu", "route.breaker.clears"},
				4: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMarketGiftController) Handle(player *pb.Player, update tgbotapi.Update) {
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
func (c *SafePlanetMarketGiftController) Validator() (hasErrors bool) {
	var err error

	switch c.CurrentState.Stage {
	case 1:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.gift.resources") {
			c.Payload.ItemType = "resources"
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.gift.items") {
			c.Payload.ItemType = "items"
		}

	// ##################################################################################################
	// Verifico se il player possiete l'item passato
	// ##################################################################################################
	case 2:
		// Verifico se il player possiede la risorsa passata
		var haveResource bool

		// Recupero nome item che il player vuole usare
		var itemChoosed string
		itemSplit := strings.Split(c.Update.Message.Text, " (")
		if len(itemSplit)-1 > 0 {
			itemSplit = strings.Split(itemSplit[0], "- ")
			if len(itemSplit)-1 > 0 {
				itemChoosed = itemSplit[1]
			}
		}

		switch c.Payload.ItemType {
		case "resources":
			// Recupero tutte le risorse del player
			var rGetPlayerResources *pb.GetPlayerResourcesResponse
			if rGetPlayerResources, err = config.App.Server.Connection.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			for _, resource := range rGetPlayerResources.GetPlayerInventory() {
				if resource.GetResource().GetName() == itemChoosed && resource.GetQuantity() > 0 {
					c.Payload.ItemID = resource.GetResource().GetID()
					haveResource = true
				}
			}
		case "items":

			var rGetPlayerItems *pb.GetPlayerItemsResponse
			if rGetPlayerItems, err = config.App.Server.Connection.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			for _, item := range rGetPlayerItems.GetPlayerInventory() {
				if helpers.Trans(c.Player.Language.Slug, "items."+item.GetItem().GetSlug()) == itemChoosed && item.GetQuantity() > 0 {
					c.Payload.ItemID = item.GetItem().GetID()
					haveResource = true
				}
			}
		}

		if !haveResource {
			return true
		}

	// ##################################################################################################
	// Verifico se il player scelto esisteo
	// ##################################################################################################
	case 3:
		var rGetPlayerByUsername *pb.GetPlayerByUsernameResponse
		rGetPlayerByUsername, _ = config.App.Server.Connection.GetPlayerByUsername(helpers.NewContext(1), &pb.GetPlayerByUsernameRequest{
			Username: c.Update.Message.Text,
		})

		if rGetPlayerByUsername.GetPlayer().GetID() <= 0 || c.Update.Message.Text == "" {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.protectors.protectors_player_not_exists")
			return true
		}

		c.Payload.ToPlayerID = rGetPlayerByUsername.GetPlayer().GetID()
		c.Payload.ToPlayerChatID = rGetPlayerByUsername.GetPlayer().GetChatID()
		c.Payload.Username = c.Update.Message.Text

	// ##################################################################################################
	// Verifico conferma player
	// ##################################################################################################
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
func (c *SafePlanetMarketGiftController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Faccio scegliere la categoria
	// ##################################################################################################
	case 0:
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.gift.intro"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.gift.resources")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.gift.items")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		// Invio messaggio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	case 1:
		// Costruisco keyboard
		var keyboardRow [][]tgbotapi.KeyboardButton

		// Recupero item per la categoria scelta
		switch c.Payload.ItemType {
		case "resources":
			// Recupero tutte le risorse del player
			var rGetPlayerResources *pb.GetPlayerResourcesResponse
			if rGetPlayerResources, err = config.App.Server.Connection.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			for i, resource := range rGetPlayerResources.GetPlayerInventory() {
				if resource.GetQuantity() > 0 && i <= 200 {
					keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							fmt.Sprintf(
								"%s - %s (%s) (%v) %s\n",
								helpers.GetResourceCategoryIcons(resource.GetResource().GetResourceCategoryID()),
								resource.GetResource().GetName(),
								strings.ToUpper(resource.GetResource().GetRarity().GetSlug()),
								resource.GetQuantity(),
								helpers.GetResourceBaseIcons(resource.GetResource().GetBase()),
							),
						),
					))
				}
			}

		case "items":
			// Recupero tutti gli item del player
			var rGetPlayerItems *pb.GetPlayerItemsResponse
			if rGetPlayerItems, err = config.App.Server.Connection.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			for _, item := range rGetPlayerItems.GetPlayerInventory() {
				if item.GetQuantity() > 0 {
					keyboardRow = append(keyboardRow,
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton(
								fmt.Sprintf(
									"%s - %s (%v)",
									helpers.Trans(c.Player.Language.Slug, "items."+item.Item.ItemCategory.Slug),
									helpers.Trans(c.Player.Language.Slug, "items."+item.Item.Slug),
									item.Quantity,
								),
							),
						),
					)
				}
			}
		}

		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		// Mando messaggio
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.gift.which"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}

		// Invio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		// Aggiungo torna al menu
		var protectorsKeyboard [][]tgbotapi.KeyboardButton
		protectorsKeyboard = append(protectorsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.gift.who"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       protectorsKeyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3

	case 3:
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.gift.confirm", c.GetRecapItem(), c.Payload.Username))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 4
	case 4:
		var giftType pb.GiftRequest_GiftTypeEnum
		switch c.Payload.ItemType {
		case "resources":
			giftType = pb.GiftRequest_RESOURCE
		case "items":
			giftType = pb.GiftRequest_ITEM
		}

		// Richiamo ws per effettuare scambio
		if _, err = config.App.Server.Connection.Gift(helpers.NewContext(1), &pb.GiftRequest{
			FromPlayerID: c.Player.ID,
			ToPlayerID:   c.Payload.ToPlayerID,
			GiftType:     giftType,
			ItemID:       c.Payload.ItemID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Se no ci sono stati errori confermo esito invio
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.gift.ok"))
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio al ricevente
		msgToReciver := helpers.NewMessage(c.Payload.ToPlayerChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.gift.ok_to_reciver", c.GetRecapItem(), c.Player.GetUsername()))
		msgToReciver.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msgToReciver); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Completed = true
	}

}

func (c *SafePlanetMarketGiftController) GetRecapItem() (itemRecap string) {
	var err error

	// Recupero item che si vuole inviare
	switch c.Payload.ItemType {
	case "resources":
		var rGetResourceByID *pb.GetResourceByIDResponse
		if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
			ID: c.Payload.ItemID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		itemRecap = fmt.Sprintf("%s %s (%s) %s",
			helpers.GetResourceCategoryIcons(rGetResourceByID.GetResource().GetResourceCategoryID()),
			rGetResourceByID.GetResource().GetName(),
			strings.ToUpper(rGetResourceByID.GetResource().GetRarity().GetSlug()),
			helpers.GetResourceBaseIcons(rGetResourceByID.GetResource().GetBase()),
		)
	case "items":
		var rGetItemByID *pb.GetItemByIDResponse
		if rGetItemByID, err = config.App.Server.Connection.GetItemByID(helpers.NewContext(1), &pb.GetItemByIDRequest{
			ItemID: c.Payload.ItemID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		itemRecap = helpers.Trans(c.Player.Language.Slug, "items."+rGetItemByID.GetItem().GetSlug())
	}

	return
}
