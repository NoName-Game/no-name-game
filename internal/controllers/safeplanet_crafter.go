package controllers

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetCrafterController (NPC Crafter)
// ====================================
type SafePlanetCrafterController struct {
	Payload struct {
		ItemType     string
		ItemCategory string
		Resources    map[uint32]int32
		AddResource  bool // Flag per verifica aggiunta nuova risorsa
		Price        int32
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetCrafterController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.crafter",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
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
	c.Stage()

	// Completo progressione
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetCrafterController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		return false
	case 1:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "armor") {
			c.Payload.ItemType = "armor"
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "weapon") {
			// Se viene richiesto di craftare un'arma passo direttamente alla lista delle risorse
			// in quanto le armi non hanno una categoria
			c.CurrentState.Stage = 2
			c.Payload.ItemType = "weapon"

			return false
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true
	case 2:
		if c.Payload.ItemCategory = helpers.CheckAndReturnCategorySlug(c.Player.Language.Slug, c.Update.Message.Text); c.Payload.ItemCategory != "" {
			return false
		}
		return true
	case 3:
		if strings.Contains(c.Update.Message.Text, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.add")) {
			// Il player ha aggiunto una nuova risorsa
			c.CurrentState.Stage = 2
			c.Payload.AddResource = true

			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.start") {
			if len(c.Payload.Resources) > 0 {
				return false
			}
		}
	case 4:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "confirm") {
			// TODO: da spostare su ws
			// Verifico se il player ha i soldi per pagare il lavoro
			var rGetPlayerEconomyMoney *pb.GetPlayerEconomyResponse
			if rGetPlayerEconomyMoney, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
				PlayerID:    c.Player.GetID(),
				EconomyType: pb.GetPlayerEconomyRequest_MONEY,
			}); err != nil {
				c.Logger.Panic(err)
			}

			if rGetPlayerEconomyMoney.GetValue() < int32(c.Payload.Price) {
				c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_money")
				return true
			}

			return false
		}
	case 5:
		var rCrafterCheck *pb.CrafterCheckResponse
		if rCrafterCheck, err = config.App.Server.Connection.CrafterCheck(helpers.NewContext(1), &pb.CrafterCheckRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Il crafter sta già portando a terminre un lavoro per questo player
		if !rCrafterCheck.GetFinishCrafting() {
			var finishAt time.Time
			if finishAt, err = helpers.GetEndTime(rCrafterCheck.GetCraftingEndTime(), c.Player); err != nil {
				c.Logger.Panic(err)
			}

			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"safeplanet.crafting.wait_validation",
				finishAt.Format("15:04:05"),
			)

			return true
		}

		return false
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetCrafterController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// Invio messaggio con recap stats
	case 0:
		startMsg := fmt.Sprintf("%s %s",
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.what"),
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.info"),
		)

		msg := helpers.NewMessage(c.Player.ChatID, startMsg)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "armor")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "weapon")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		// Invio messaggio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1
	case 1:
		var message string
		var keyboardRowCategories [][]tgbotapi.KeyboardButton

		switch c.Payload.ItemType {
		case "armor":
			message = helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.armor.type")

			var rGetAllArmorCategory *pb.GetAllArmorCategoryResponse
			if rGetAllArmorCategory, err = config.App.Server.Connection.GetAllArmorCategory(helpers.NewContext(1), &pb.GetAllArmorCategoryRequest{}); err != nil {
				c.Logger.Panic(err)
			}

			for _, category := range rGetAllArmorCategory.GetArmorCategories() {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, category.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		msg := helpers.NewMessage(c.Player.ChatID, message)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}

		// Invio messaggio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		var rGetPlayerResources *pb.GetPlayerResourcesResponse
		if rGetPlayerResources, err = config.App.Server.Connection.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Se l'inventario è vuoto allora concludi
		if len(rGetPlayerResources.GetPlayerInventory()) <= 0 {
			message := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_resources"))
			message.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
					),
				),
			)

			if _, err = helpers.SendMessage(message); err != nil {
				c.Logger.Panic(err)
			}
			// Completo lo stato
			c.CurrentState.Completed = true
		}

		type CraftResourceStruct struct {
			ResourceName   string
			ResourceRarity string
			ResourceID     uint32
			Quantity       int32
		}

		// Mappo tutte le risorse del player
		var playerResources []CraftResourceStruct
		for _, resource := range rGetPlayerResources.GetPlayerInventory() {
			playerResources = append(playerResources, CraftResourceStruct{
				ResourceID:     resource.GetResource().GetID(),
				ResourceName:   resource.GetResource().GetName(),
				ResourceRarity: resource.GetResource().GetRarity().GetSlug(),
				Quantity:       resource.GetQuantity(),
			})
		}

		// Se è stato aggiunto una risorsa ovvero quando viene processto il messaggio "aggiungi"
		if c.Payload.AddResource {
			// Se è la prima risorsa inizializzo la mappa
			if c.Payload.Resources == nil {
				c.Payload.Resources = make(map[uint32]int32)
			}

			// Recupero risorsa da messaggio
			resourceName := strings.Split(
				strings.Split(c.Update.Message.Text, " (")[0],
				helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.add")+" ",
			)[1]

			// Recupero risorsa
			var rGetResourceByName *pb.GetResourceByNameResponse
			if rGetResourceByName, err = config.App.Server.Connection.GetResourceByName(helpers.NewContext(1), &pb.GetResourceByNameRequest{
				Name: resourceName,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Recupero dettagli risorsa
			choosedResource := rGetResourceByName.GetResource()

			// Controllo che l'utente abbia effettivamente l'item
			hasResource := false
			for _, resource := range playerResources {
				if resource.ResourceID == choosedResource.GetID() {
					hasResource = true

					// TODO: spostare questa logica sul ws
					// Aumento prezzo in base alla rarità della risorsa usata
					c.Payload.Price += int32(10 * choosedResource.GetRarity().GetID())

					// Se il player ha effettivamente la risorsa creo/incremento
					// Incremento quantitativo risorse
					if helpers.KeyInMap(choosedResource.GetID(), c.Payload.Resources) && hasResource {
						if c.Payload.Resources[choosedResource.GetID()] < resource.Quantity {
							c.Payload.Resources[choosedResource.GetID()]++
						}
					} else if hasResource {
						c.Payload.Resources[choosedResource.GetID()] = 1
					}
				}
			}

			// Risorsa non trovata! invio errore
			if !hasResource {
				msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_resource"))
				if _, err = helpers.SendMessage(msg); err != nil {
					c.Logger.Panic(err)
				}
			}
		}

		// Costruisco keyboard
		var keyboardRowResources [][]tgbotapi.KeyboardButton

		// Se sono già stati inseriti delle risorse mostro tasto start craft!
		if len(c.Payload.Resources) > 0 {
			keyboardRowResources = append(keyboardRowResources, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.start"),
				),
			))
		}

		// Inserisco lista delle risorse
		for _, resource := range playerResources {
			if c.Payload.Resources[resource.ResourceID] <= resource.Quantity {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
					fmt.Sprintf("%s %s (%s) %v/%v",
						helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.add"),
						resource.ResourceName,
						resource.ResourceRarity,
						resource.Quantity-c.Payload.Resources[resource.ResourceID], resource.Quantity,
					),
				))
				keyboardRowResources = append(keyboardRowResources, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowResources = append(keyboardRowResources,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)

		// Add recipe message
		var recipe string
		if len(c.Payload.Resources) > 0 {
			for resourceID, quantity := range c.Payload.Resources {
				var rGetResourceByID *pb.GetResourceByIDResponse
				if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
					ID: resourceID,
				}); err != nil {
					c.Logger.Panic(err)
				}

				recipe += fmt.Sprintf("- *%v* x %s (%s)\n",
					quantity,
					rGetResourceByID.GetResource().Name,
					rGetResourceByID.GetResource().GetRarity().GetSlug(),
				)
			}
		}

		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
			"safeplanet.crafting.choose_resources",
			c.Payload.Price,
			recipe,
		))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowResources,
		}

		// Invio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3
	case 3:
		// =========================
		// Recap risorse usate per il crafting
		// =========================
		var recipe string
		if len(c.Payload.Resources) > 0 {
			for resourceID, quantity := range c.Payload.Resources {
				var rGetResourceByID *pb.GetResourceByIDResponse
				if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
					ID: resourceID,
				}); err != nil {
					c.Logger.Panic(err)
				}

				recipe += fmt.Sprintf("- *%v* x %s (%s)\n",
					quantity,
					rGetResourceByID.GetResource().Name,
					rGetResourceByID.GetResource().GetRarity().GetSlug(),
				)
			}
		}

		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
			"safeplanet.crafting.confirm_choose_resources",
			c.Payload.Price,
			recipe,
		))
		msg.ParseMode = "markdown"
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
		c.CurrentState.Stage = 4
	case 4:
		// =========================
		// Start crating
		// =========================
		var rCrafterStart *pb.CrafterStartResponse
		if rCrafterStart, err = config.App.Server.Connection.CrafterStart(helpers.NewContext(1), &pb.CrafterStartRequest{
			PlayerID:     c.Player.ID,
			Resources:    c.Payload.Resources,
			Price:        c.Payload.Price,
			ItemType:     c.Payload.ItemType,
			ItemCategory: c.Payload.ItemCategory,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Converto finishAt in formato Time
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rCrafterStart.GetCraftingEndTime(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		var msg tgbotapi.MessageConfig
		msg = helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.wait", finishAt.Format("15:04:05")),
		)
		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 5
		c.ForceBackTo = true
	case 5:
		var rCrafterEnd *pb.CrafterEndResponse
		if rCrafterEnd, err = config.App.Server.Connection.CrafterEnd(helpers.NewContext(1), &pb.CrafterEndRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		var endCraftMessage string
		if rCrafterEnd.GetArmor() != nil {
			endCraftMessage = helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.craft_completed", rCrafterEnd.GetArmor().GetName())
		} else if rCrafterEnd.GetWeapon() != nil {
			endCraftMessage = helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.craft_completed", rCrafterEnd.GetWeapon().GetName())
		}

		msg := helpers.NewMessage(c.Player.ChatID, endCraftMessage)
		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
