package controllers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetCrafterController (NPC Crafter)
// ====================================
type SafePlanetCrafterController struct {
	Payload struct {
		Item      string
		Category  string
		Resources map[uint32]int32
		AddFlag   bool
		Price     uint32
	}
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetCrafterController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller: "route.safeplanet.crafter",
		Payload:    c.Payload,
		ControllerBack: ControllerBack{
			To:        &MenuController{},
			FromStage: 0,
		},
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
func (c *SafePlanetCrafterController) Validator() (hasErrors bool) {
	var err error
	switch c.PlayerData.CurrentState.Stage {
	case 0:
		return false
		// nothinggg
	case 1:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "armors") {
			c.Payload.Item = "armors"
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "weapon") {
			// Se viene richiesto di craftare un'arma passo direttamente alla lista delle risorse
			// in quanto le armi non hanno una categoria
			c.PlayerData.CurrentState.Stage = 2
			c.Payload.Item = "weapon"

			return false
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true
	case 2:
		if c.Payload.Category = helpers.CheckAndReturnCategorySlug(c.Player.Language.Slug, c.Update.Message.Text); c.Payload.Category != "" {
			return false
		}
		return true
	case 3:
		if strings.Contains(c.Update.Message.Text, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.add")) {
			c.PlayerData.CurrentState.Stage = 2
			c.Payload.AddFlag = true
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.start") {
			if len(c.Payload.Resources) > 0 {
				return false
			}
		}
	case 4:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "confirm") {
			// Verifico se il player ha i soldi per pagare il lavoro
			var rGetPlayerEconomyMoney *pb.GetPlayerEconomyResponse
			rGetPlayerEconomyMoney, err = services.NnSDK.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
				PlayerID:    c.Player.GetID(),
				EconomyType: "money",
			})
			if err != nil {
				panic(err)
			}

			if rGetPlayerEconomyMoney.GetValue() < int32(c.Payload.Price) {
				c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_money")
				return true
			}

			return false
		}
	case 5:
		var finishAt time.Time
		finishAt, err = ptypes.Timestamp(c.PlayerData.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"safeplanet.crafting.wait_validation",
			finishAt.Format("15:04:05"),
		)

		// Verifico se ha finito il crafting
		finishAt, err = ptypes.Timestamp(c.PlayerData.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		if time.Now().After(finishAt) {
			return false
		}
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetCrafterController) Stage() (err error) {
	switch c.PlayerData.CurrentState.Stage {
	// Invio messaggio con recap stats
	case 0:
		startMsg := fmt.Sprintf("%s %s",
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.what"),
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.info"),
		)

		msg := services.NewMessage(c.Player.ChatID, startMsg)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "armors")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "weapon")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}
		// Avanzo di stage
		c.PlayerData.CurrentState.Stage = 1
	case 1:
		var message string
		var keyboardRowCategories [][]tgbotapi.KeyboardButton

		switch c.Payload.Item {
		case "armors":
			message = helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.armor.type")

			var rGetAllArmorCategory *pb.GetAllArmorCategoryResponse
			rGetAllArmorCategory, err = services.NnSDK.GetAllArmorCategory(helpers.NewContext(1), &pb.GetAllArmorCategoryRequest{})
			if err != nil {
				return err
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

		msg := services.NewMessage(c.Player.ChatID, message)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.PlayerData.CurrentState.Stage = 2
	case 2:
		var rGetPlayerResources *pb.GetPlayerResourcesResponse
		rGetPlayerResources, err = services.NnSDK.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
			PlayerID: c.Player.ID,
		})
		if err != nil {
			return err
		}

		// Se l'inventario è vuoto allora concludi
		if len(rGetPlayerResources.GetPlayerInventory()) <= 0 {
			message := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_resources"))
			message.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
					),
				),
			)
			_, err = services.SendMessage(message)
			if err != nil {
				return err
			}
			// Completo lo stato
			c.PlayerData.CurrentState.Completed = true
		}

		type CraftResourceStruct struct {
			ResourceID     uint32
			ResourceName   string
			ResourceRarity string
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
		if c.Payload.AddFlag {
			// Se è la prima risorsa inizializzo la mappa
			if c.Payload.Resources == nil {
				c.Payload.Resources = make(map[uint32]int32)
			}

			// Recupero risorsa da messaggio
			resourceName := strings.Split(
				strings.Split(c.Update.Message.Text, " (")[0],
				helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.add")+" ",
			)[1]

			var rGetResourceByName *pb.GetResourceByNameResponse
			rGetResourceByName, err = services.NnSDK.GetResourceByName(helpers.NewContext(1), &pb.GetResourceByNameRequest{
				Name: resourceName,
			})
			if err != nil {
				return err
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
					c.Payload.Price += 10 * choosedResource.GetRarity().GetID()

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

			// Se non è stato trovata corrispondenza invio errore
			if !hasResource {
				msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_resource"))
				_, err = services.SendMessage(msg)
				if err != nil {
					return err
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
				rGetResourceByID, err = services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
					ID: resourceID,
				})
				if err != nil {
					return err
				}

				recipe += fmt.Sprintf("- *%v* x %s (%s)\n",
					quantity,
					rGetResourceByID.GetResource().Name,
					rGetResourceByID.GetResource().GetRarity().GetSlug(),
				)
			}
		}

		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
			"safeplanet.crafting.choose_resources",
			c.Payload.Price,
			recipe,
		))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowResources,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.PlayerData.CurrentState.Stage = 3
	case 3:
		// Add recipe message
		var recipe string
		if len(c.Payload.Resources) > 0 {
			for resourceID, quantity := range c.Payload.Resources {
				var rGetResourceByID *pb.GetResourceByIDResponse
				rGetResourceByID, err = services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
					ID: resourceID,
				})
				if err != nil {
					return err
				}

				recipe += fmt.Sprintf("- *%v* x %s (%s)\n",
					quantity,
					rGetResourceByID.GetResource().Name,
					rGetResourceByID.GetResource().GetRarity().GetSlug(),
				)
			}
		}

		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
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
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.PlayerData.CurrentState.Stage = 4
	case 4:
		// Il player ha avviato il crafting, Rimuovo risorse usate al player
		for resourceID, quantity := range c.Payload.Resources {
			_, err := services.NnSDK.ManagePlayerInventory(helpers.NewContext(1), &pb.ManagePlayerInventoryRequest{
				PlayerID: c.Player.GetID(),
				ItemID:   resourceID,
				ItemType: "resources",
				Quantity: -quantity,
			})
			if err != nil {
				return err
			}
		}

		// Rimuovo money
		_, err = services.NnSDK.CreateTransaction(helpers.NewContext(1), &pb.CreateTransactionRequest{
			Value:                 -int32(c.Payload.Price),
			TransactionTypeID:     1, // Gold
			TransactionCategoryID: 9, // Crafter Safe Planet
			PlayerID:              c.Player.GetID(),
		})

		// Stampa il tempo di attesa e aggiorna stato
		endTime := helpers.GetEndTime(0, 10, 0)
		var msg tgbotapi.MessageConfig
		msg = services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.wait", endTime.Format("15:04:05")),
		)
		msg.ParseMode = "markdown"
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		endTimeProto, err := ptypes.TimestampProto(endTime)
		if err != nil {
			panic(err)
		}

		c.PlayerData.CurrentState.FinishAt = endTimeProto
		c.PlayerData.CurrentState.ToNotify = true
		c.PlayerData.CurrentState.Stage = 5
		c.ForceBackTo = true
	case 5:
		// crafting completato
		items, err := json.Marshal(c.Payload.Resources)
		if err != nil {
			return err
		}

		var text string
		switch c.Payload.Item {
		case "armors":
			// Creo la richiesta di craft armor
			var rCraftArmor *pb.CraftArmorResponse
			rCraftArmor, err = services.NnSDK.CraftArmor(helpers.NewContext(1), &pb.CraftArmorRequest{
				Category: c.Payload.Category,
				Items:    string(items),
				PlayerID: c.Player.ID,
			})
			if err != nil {
				return err
			}

			text = helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.craft_completed", rCraftArmor.GetArmor().GetName())
		case "weapon":
			// Creo la richiesta di craft weapon
			var rCraftWeapon *pb.CraftWeaponResponse
			rCraftWeapon, err = services.NnSDK.CraftWeapon(helpers.NewContext(1), &pb.CraftWeaponRequest{
				Items:    string(items),
				PlayerID: c.Player.ID,
			})
			if err != nil {
				return err
			}

			text = helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.craft_completed", rCraftWeapon.GetWeapon().GetName())
		}

		msg := services.NewMessage(c.Player.ChatID, text)
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
