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
// SafePlanetResourceBankController
// ====================================
type SafePlanetResourceBankController struct {
	Payload struct {
		Type       string
		ResourceID uint32
		Offset     uint32
	}
	Controller
}

func (c *SafePlanetResourceBankController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.bank.resources",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				1: {"route.breaker.menu", "route.breaker.back"},
				2: {"route.breaker.back"},
				3: {"route.breaker.back"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetResourceBankController) Handle(player *pb.Player, update tgbotapi.Update) {
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
func (c *SafePlanetResourceBankController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico tipologia di transazione
	// ##################################################################################################
	case 1:
		/*if helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit"),
			helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws"),
		}) {
			c.CurrentState.Stage = 1
		}*/
		switch c.Update.Message.Text {
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit"):
			c.Payload.Type = "deposit"
			c.CurrentState.Stage = 1
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws"):
			c.Payload.Type = "withdraws"
			c.CurrentState.Stage = 1
		}
	case 2:
		switch c.Update.Message.Text {
		case helpers.Trans(c.Player.Language.Slug, "back"):
			if c.Payload.Offset > 0 {
				c.Payload.Offset--
			}
			c.CurrentState.Stage = 1
			return false
		case helpers.Trans(c.Player.Language.Slug, "next"):
			c.Payload.Offset++
			c.CurrentState.Stage = 1
			return false
		}
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.all") && c.Payload.Type == "deposit" {
			c.CurrentState.Stage = 3
			return false
		}
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

		var playerInventories []*pb.PlayerInventory
		switch c.Payload.Type {
		case "deposit":
			// Recupero tutte le risorse del player
			var rGetPlayerResources *pb.GetPlayerResourcesResponse
			if rGetPlayerResources, err = config.App.Server.Connection.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}
			playerInventories = rGetPlayerResources.GetPlayerInventory()
		case "withdraws":
			// Recupero Risorse e le metto
			var rGetDepositedResources *pb.GetDepositedResourcesResponse
			if rGetDepositedResources, err = config.App.Server.Connection.GetDepositedResources(helpers.NewContext(1), &pb.GetDepositedResourcesRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Error(err.Error())
			}
			playerInventories = rGetDepositedResources.GetPlayerInventory()
		}

		for _, resource := range playerInventories {
			if resource.GetResource().GetName() == itemChoosed && resource.GetQuantity() > 0 {
				c.Payload.ResourceID = resource.GetResource().GetID()
				haveResource = true
			}
		}

		if !haveResource {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetResourceBankController) Stage() {

	var err error
	switch c.CurrentState.Stage {
	// Invio messaggio con recap stats
	case 0:
		var rGetDepositedResources *pb.GetDepositedResourcesResponse
		if rGetDepositedResources, err = config.App.Server.Connection.GetDepositedResources(helpers.NewContext(1), &pb.GetDepositedResourcesRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Error(err.Error())
		}

		var nResource int32
		for _, r := range rGetDepositedResources.GetPlayerInventory() {
			if r.GetQuantity() > 0 {
				nResource += r.GetQuantity()
			}
		}

		msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s\n\n%s",
			helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.info"),
			helpers.Trans(
				c.Player.Language.Slug,
				"safeplanet.bank.resource.account_details",
				nResource,
			),
		))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1
	case 1:
		var keyboardRow [][]tgbotapi.KeyboardButton
		var playerInventories []*pb.PlayerInventory
		switch c.Payload.Type {
		case "deposit":
			// Recupero tutte le risorse del player
			var rGetPlayerResources *pb.GetPlayerResourcesResponse
			if rGetPlayerResources, err = config.App.Server.Connection.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}
			playerInventories = rGetPlayerResources.GetPlayerInventory()
			// Aggiungo alla tastiera un tasto per recuperare tutto
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.all"))))
		case "withdraws":
			// Recupero Risorse e le metto
			var rGetDepositedResources *pb.GetDepositedResourcesResponse
			if rGetDepositedResources, err = config.App.Server.Connection.GetDepositedResources(helpers.NewContext(1), &pb.GetDepositedResourcesRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Error(err.Error())
			}
			playerInventories = rGetDepositedResources.GetPlayerInventory()
		}
		var start, end int
		if start = int(c.Payload.Offset) * 50; start >= len(playerInventories) {
			start = 0
		}
		if end = start + 50; end >= len(playerInventories) {
			end = len(playerInventories)
		}
		for _, resource := range playerInventories[start:end] {
			if resource.GetQuantity() > 0 {
				keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						fmt.Sprintf(
							"%s - %s (%s) (%v) %s",
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

		if len(playerInventories) > 50 {
			// appendo i bottoni next e back per cambiare l'offset
			var row []tgbotapi.KeyboardButton
			if c.Payload.Offset > 0 {
				row = append(row, tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "back")))
			}
			row = append(row, tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "next")))
			keyboardRow = append(keyboardRow, row)
		}

		// Aggiungo tasti back and clears
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.what_resource"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		var mainMessage string
		var keyboardRowQuantities [][]tgbotapi.KeyboardButton
		switch c.Payload.Type {
		case "deposit":
			mainMessage = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit_message")
		case "withdraws":
			mainMessage = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws_message")
		}
		// Inserisco le quantità di default per il prelievo/deposito
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
		msg := helpers.NewMessage(c.ChatID, mainMessage)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowQuantities,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}
		c.CurrentState.Stage = 4
	case 3:
		// DEPOSITO TUTTO
		var text string
		text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.operation_done")

		if _, err = config.App.Server.Connection.DepositAllResources(helpers.NewContext(1), &pb.DepositAllResourcesRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			if strings.Contains(err.Error(), "no resources in inventory") {
				text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.resource.transaction_error")
			} else {
				c.Logger.Panic(err)
			}
		}
		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID, text)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
		c.Configurations.ControllerBack.To = &SafePlanetResourceBankController{}
	case 4:
		// Se la validazione è passata vuol dire che è stato
		// inserito un importo valido e quindi posso eseguiore la transazione
		// in base alla tipologia scelta

		// Converto valore richiesto in int
		var value int
		if value, err = strconv.Atoi(c.Update.Message.Text); err != nil {
			c.Logger.Panic(err)
		}

		var text string
		text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.operation_done")

		switch c.Payload.Type {
		case "deposit":
			if _, err = config.App.Server.Connection.DepositResource(helpers.NewContext(1), &pb.DepositResourceRequest{
				ResourceID: c.Payload.ResourceID,
				PlayerID:   c.Player.ID,
				Quantity:   int32(value),
			}); err != nil {
				if strings.Contains(err.Error(), "inventory full") {
					text = helpers.Trans(c.Player.Language.Slug, "inventory.inventory_full")
				} else {
					text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.resource.transaction_error")
				}
			}
		case "withdraws":
			if _, err = config.App.Server.Connection.WithdrawResource(helpers.NewContext(1), &pb.WithdrawResourceRequest{
				ResourceID: c.Payload.ResourceID,
				PlayerID:   c.Player.ID,
				Quantity:   int32(value),
			}); err != nil {
				if strings.Contains(err.Error(), "inventory full") {
					text = helpers.Trans(c.Player.Language.Slug, "inventory.inventory_full")
				} else {
					text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.resource.transaction_error")
				}
			}
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID, text)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
		c.Configurations.ControllerBack.To = &SafePlanetResourceBankController{}
	}

	return
}
