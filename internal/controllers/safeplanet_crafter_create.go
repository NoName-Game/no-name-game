package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetCrafterCreateController (NPC Crafter)
// ====================================
type SafePlanetCrafterCreateController struct {
	Payload struct {
		ItemType         string
		ItemCategory     string
		Resources        map[uint32]int32
		ResourceQuantity int32 // Quantità di risorse aggiunte
		SingleQuantity   int32 // Quantità per singolo item
		ResourceName     string
		AddResource      bool // Flag per verifica aggiunta nuova risorsa
		Price            int32
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetCrafterCreateController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.crafter.create",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCrafterController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
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
func (c *SafePlanetCrafterCreateController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico tipologia item che il player vuole craftare
	// ##################################################################################################
	case 0:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "armor") {
			c.Payload.ItemType = "armor"
			c.CurrentState.Stage = 1
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "weapon") {
			// Se viene richiesto di craftare un'arma passo direttamente alla lista delle risorse
			// in quanto le armi non hanno una categoria
			c.Payload.ItemType = "weapon"
			c.CurrentState.Stage = 2
		}
	// ##################################################################################################
	// Verifico che il player abbia scelto una tipologia valida
	// ##################################################################################################
	case 2:
		if c.Payload.ItemCategory = helpers.CheckAndReturnCategorySlug(c.Player.Language.Slug, c.Update.Message.Text); c.Payload.ItemCategory == "" {
			return true
		}
	// ##################################################################################################
	// Verifico se il player ha deciso di inserire un nuovo elemento al craft, o di concludere l'operazione
	// ##################################################################################################
	case 3:
		if strings.Contains(c.Update.Message.Text, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.add")) {
			// Il player ha aggiunto una nuova risorsa
			c.Payload.AddResource = true

			// Recupero risorsa da messaggio, e se non rispecchia le specifiche ritorno errore
			resourceName := strings.Split(strings.Split(c.Update.Message.Text, " (")[0], " ")
			if len(resourceName) < 3 {
				return true
			} else {
				c.Payload.ResourceName = resourceName[2]
			}

		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.start") {
			// Non è possibile iniziare il craft senza risorse
			if len(c.Payload.Resources) <= 0 {
				return true
			}
			c.CurrentState.Stage = 4
		} else {
			// Se non è nessuno di questi allora ritorno errore
			return true
		}
	// ##################################################################################################
	// Verifico quantità di item che il player vuole utilizzare
	// ##################################################################################################
	case 4:
		if quantity, err := strconv.Atoi(c.Update.Message.Text); err != nil && quantity <= 0 {
			return true
		} else {
			c.Payload.SingleQuantity = int32(quantity)
			c.CurrentState.Stage = 2
		}

	// ##################################################################################################
	// Se il player ha dato conferma verifico se ha il denaro necessario per proseguire
	// ##################################################################################################
	case 5:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	// ##################################################################################################
	// Verifico stato crafting
	// ##################################################################################################
	case 6:
		var err error
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
				finishAt.Format("15:04"),
			)

			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.continue"),
					),
				),
			)

			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetCrafterCreateController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// Invio messaggio con recap stats
	case 0:
		startMsg := fmt.Sprintf("%s\n\n%s",
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.what"),
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.info"),
		)

		msg := helpers.NewMessage(c.Player.ChatID, startMsg)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "armor")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "weapon")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		// Invio messaggio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

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
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
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
			ResourceName       string
			ResourceRarity     string
			ResourceCategoryID uint32
			ResourceBase       bool
			ResourceID         uint32
			Quantity           int32
		}

		// Mappo tutte le risorse del player
		var playerResources []CraftResourceStruct
		for _, resource := range rGetPlayerResources.GetPlayerInventory() {
			playerResources = append(playerResources, CraftResourceStruct{
				ResourceID:         resource.GetResource().GetID(),
				ResourceName:       resource.GetResource().GetName(),
				ResourceRarity:     resource.GetResource().GetRarity().GetSlug(),
				ResourceCategoryID: resource.GetResource().GetResourceCategoryID(),
				Quantity:           resource.GetQuantity(),
				ResourceBase:       resource.GetResource().GetBase(),
			})
		}

		// Se è stato aggiunto una risorsa ovvero quando viene processto il messaggio "aggiungi"
		if c.Payload.AddResource {
			// Se è la prima risorsa inizializzo la mappa
			if c.Payload.Resources == nil {
				c.Payload.Resources = make(map[uint32]int32)
			}

			// Recupero risorsa
			var rGetResourceByName *pb.GetResourceByNameResponse
			if rGetResourceByName, err = config.App.Server.Connection.GetResourceByName(helpers.NewContext(1), &pb.GetResourceByNameRequest{
				Name: c.Payload.ResourceName,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Recupero dettagli risorsa
			choosedResource := rGetResourceByName.GetResource()

			// Controllo che l'utente abbia effettivamente l'item
			hasResource := false
			for _, resource := range playerResources {
				if resource.ResourceID == choosedResource.GetID() {
					// Controllo che il player abbia effettivamente la quantità richiesta.
					if resource.Quantity >= c.Payload.SingleQuantity {
						hasResource = true
						// Se il player ha effettivamente la risorsa creo/incremento
						// Incremento quantitativo risorse
						if helpers.KeyInMap(choosedResource.GetID(), c.Payload.Resources) && hasResource {
							if c.Payload.Resources[choosedResource.GetID()]+c.Payload.SingleQuantity < resource.Quantity {
								c.Payload.Resources[choosedResource.GetID()] += c.Payload.SingleQuantity
								c.Payload.Price += int32(10*choosedResource.GetRarity().GetID()) * c.Payload.SingleQuantity
							}
						} else if hasResource {
							// È la prima volta che inserisce questa risorsa
							c.Payload.Resources[choosedResource.GetID()] = c.Payload.SingleQuantity
							c.Payload.Price += int32(10*choosedResource.GetRarity().GetID()) * c.Payload.SingleQuantity
						}

						c.Payload.ResourceQuantity += c.Payload.SingleQuantity
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
		if c.Payload.ResourceQuantity >= 20 {
			keyboardRowResources = append(keyboardRowResources, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.start"),
				),
			))
		}

		// Inserisco lista delle risorse
		for _, resource := range playerResources {
			if c.Payload.Resources[resource.ResourceID] <= resource.Quantity {
				// Verifico se la quantità disponibile sia sopra allo 0
				availabeQuantity := resource.Quantity - c.Payload.Resources[resource.ResourceID]
				if availabeQuantity > 0 {
					// Verifico se è una risorsa base
					baseResources := ""
					if resource.ResourceBase {
						baseResources = "🔬Base"
					}

					keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
						fmt.Sprintf("%s %s %s (%s) %v/%v %s",
							helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.add"),
							helpers.GetResourceCategoryIcons(resource.ResourceCategoryID),
							resource.ResourceName,
							resource.ResourceRarity,
							resource.Quantity-c.Payload.Resources[resource.ResourceID], resource.Quantity,
							baseResources,
						),
					))
					keyboardRowResources = append(keyboardRowResources, keyboardRow)
				}
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

				recipe += fmt.Sprintf("- <b>%v</b> x %s %s (%s) %s \n",
					quantity,
					helpers.GetResourceCategoryIcons(rGetResourceByID.GetResource().GetResourceCategoryID()),
					rGetResourceByID.GetResource().Name,
					rGetResourceByID.GetResource().GetRarity().GetSlug(),
					helpers.GetResourceBaseIcons(rGetResourceByID.GetResource().GetBase()),
				)
			}
		}

		msgContent := helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.choose_resources",
			c.Payload.Price,
			recipe,
		)

		if c.Payload.ResourceQuantity < 20 {
			msgContent += helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.min_resources",
				20-c.Payload.ResourceQuantity,
			)
		}

		msg := helpers.NewMessage(c.Player.ChatID, msgContent)
		msg.ParseMode = tgbotapi.ModeHTML
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
		// Chiedo il quantitativo di risorse che vuole utilizzare
		// =========================
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
			"safeplanet.crafting.how_many",
		))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("1"),
				tgbotapi.NewKeyboardButton("5"),
				tgbotapi.NewKeyboardButton("10"),
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

				// Verifico se è una risorsa base
				baseResources := ""
				if rGetResourceByID.GetResource().GetBase() {
					baseResources = "🔬Base"
				}

				recipe += fmt.Sprintf("- <b>%v</b> x %s %s (%s) %s \n",
					quantity,
					helpers.GetResourceCategoryIcons(rGetResourceByID.GetResource().GetResourceCategoryID()),
					rGetResourceByID.GetResource().Name,
					rGetResourceByID.GetResource().GetRarity().GetSlug(),
					baseResources,
				)
			}
		}

		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
			"safeplanet.crafting.confirm_choose_resources",
			c.Payload.Price,
			recipe,
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
		c.CurrentState.Stage = 5
	case 5:
		// =========================
		// Start crating
		// =========================
		var rCrafterStart *pb.CrafterStartResponse
		rCrafterStart, err = config.App.Server.Connection.CrafterStart(helpers.NewContext(1), &pb.CrafterStartRequest{
			PlayerID:     c.Player.ID,
			Resources:    c.Payload.Resources,
			Price:        c.Payload.Price,
			ItemType:     c.Payload.ItemType,
			ItemCategory: c.Payload.ItemCategory,
		})

		if err != nil && strings.Contains(err.Error(), "player dont have enough money") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di monete
			errorMsg := helpers.NewMessage(c.ChatID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_money"),
			)

			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		} else if err != nil {
			c.Logger.Panic(err)
		}

		// Converto finishAt in formato Time
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rCrafterStart.GetCraftingEndTime(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		var msg tgbotapi.MessageConfig
		msg = helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.wait", finishAt.Format("15:04")),
		)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 6
		c.ForceBackTo = true
	case 6:
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
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
