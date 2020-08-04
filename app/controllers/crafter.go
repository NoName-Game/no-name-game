package controllers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// CrafterController (NPC Crafter)
// ====================================
type CrafterController struct {
	Payload struct {
		Item      string
		Category  string
		Resources map[uint32]int32
		AddFlag   bool
	}
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *CrafterController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.safeplanet.crafter",
		c.Payload,
		[]string{},
		player,
		update,
	) {
		return
	}

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &MenuController{}) {
			return
		}
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.CurrentState.Payload, &c.Payload)

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
		// validatorMsg.ReplyMarkup = c.Validation.ReplyKeyboard

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
	c.CurrentState.Payload = string(payloadUpdated)

	rUpdatePlayerState, err := services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
		PlayerState: c.CurrentState,
	})
	if err != nil {
		panic(err)
	}
	c.CurrentState = rUpdatePlayerState.GetPlayerState()

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *CrafterController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		),
	)

	switch c.CurrentState.Stage {
	case 0:
		return false, err
		// nothinggg
	case 1:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "armors") {
			c.Payload.Item = "armors"
			return false, err
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "weapons") {
			c.Payload.Item = "weapons"
			return false, err
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true, err
	case 2:
		if c.Payload.Category = helpers.CheckAndReturnCategorySlug(c.Player.Language.Slug, c.Update.Message.Text); c.Payload.Category != "" {
			return false, err
		}
		return true, err
	case 3:
		if strings.Contains(c.Update.Message.Text, helpers.Trans(c.Player.Language.Slug, "crafting.add")) {
			c.CurrentState.Stage = 2
			c.Payload.AddFlag = true
			return false, err
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "crafting.start") {
			if len(c.Payload.Resources) > 0 {
				return false, err
			}
		}
	case 4:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "confirm") {
			return false, err
		}
	case 5:
		finishAt, err := ptypes.Timestamp(c.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"crafting.wait",
			finishAt.Format("15:04:05"),
		)

		// Aggiungo anche abbandona
		c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.continue"),
				),
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
				),
			),
		)

		// Verifico se ha finito il crafting
		finishAt, err = ptypes.Timestamp(c.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		if time.Now().After(finishAt) {
			return false, err
		}
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *CrafterController) Stage() (err error) {
	switch c.CurrentState.Stage {
	// Invio messaggio con recap stats
	case 0:

		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.what"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "armors")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "weapons")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}
		// Avanzo di stage
		c.CurrentState.Stage = 1
	case 1:
		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch c.Payload.Item {
		case "armors":
			rGetAllArmorCategory, err := services.NnSDK.GetAllArmorCategory(helpers.NewContext(1), &pb.GetAllArmorCategoryRequest{})
			if err != nil {
				return err
			}

			for _, category := range rGetAllArmorCategory.GetArmorCategories() {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, category.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case "weapons":
			rGetAllWeaponCategory, err := services.NnSDK.GetAllWeaponCategory(helpers.NewContext(1), &pb.GetAllWeaponCategoryRequest{})
			if err != nil {
				return err
			}

			for _, category := range rGetAllWeaponCategory.GetWeaponCategories() {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, category.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
		))

		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.type"))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}
		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		rGetPlayerResources, err := services.NnSDK.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
			PlayerID: c.Player.ID,
		})
		if err != nil {
			return err
		}

		mapInventory := helpers.InventoryResourcesToMap(rGetPlayerResources.GetPlayerInventory())

		// Se l'inventario è vuoto allora concludi
		if len(mapInventory) <= 0 {
			message := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.no_resources"))
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
			c.CurrentState.Completed = true
		}

		// Id Add new resource
		if c.Payload.AddFlag {
			if c.Payload.Resources == nil {
				c.Payload.Resources = make(map[uint32]int32)
			}

			// Clear text from Add and other shit.
			resourceName := strings.Split(
				strings.Split(c.Update.Message.Text, " (")[0],
				helpers.Trans(c.Player.Language.Slug, "crafting.add")+" ")[1]

			rGetResourceByName, err := services.NnSDK.GetResourceByName(helpers.NewContext(1), &pb.GetResourceByNameRequest{
				Name: resourceName,
			})
			if err != nil {
				return err
			}

			resourceID := rGetResourceByName.GetResource().GetID()
			resourceMaxQuantity := mapInventory[resourceID]
			hasResource := true

			// Controllo che l'utente abbia effettivamente l'item
			if mapInventory[resourceID] == 0 {
				// Non ha l'item!
				hasResource = false
				msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.no_resource"))
				_, err = services.SendMessage(msg)
				if err != nil {
					return err
				}

			}

			if helpers.KeyInMap(resourceID, c.Payload.Resources) && hasResource {
				if c.Payload.Resources[resourceID] < resourceMaxQuantity {
					c.Payload.Resources[resourceID]++
				}
			} else if hasResource {
				c.Payload.Resources[resourceID] = 1
			}
		}

		// Keyboard with resources
		var keyboardRowResources [][]tgbotapi.KeyboardButton
		for r, q := range mapInventory {
			// If PayloadResouces < Inventory quantity ok :)
			rGetResourceByID, err := services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: r,
			})
			if err != nil {
				return err
			}

			resource := rGetResourceByID.GetResource()
			if c.Payload.Resources[r] < q {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "crafting.add") + " " + resource.Name + " (" + (strconv.Itoa(int(q) - int(c.Payload.Resources[r]))) + ")",
				))
				keyboardRowResources = append(keyboardRowResources, keyboardRow)
			}
		}

		// If PayloadResources is not empty show craft button
		if len(c.Payload.Resources) > 0 {
			keyboardRowResources = append(keyboardRowResources, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "crafting.start"),
				),
			))
		}

		// Clear and exit
		keyboardRowResources = append(keyboardRowResources,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)

		// Add recipe message
		var recipe string
		if len(c.Payload.Resources) > 0 {
			for k, v := range c.Payload.Resources {
				rGetResourceByID, err := services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
					ID: k,
				})
				if err != nil {
					return err
				}

				recipe += rGetResourceByID.GetResource().Name + " x " + strconv.Itoa(int(v)) + "\n"
			}
		}

		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.choose_resources")+"\n"+recipe)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowResources,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}
		// Aggiorno stato
		c.CurrentState.Stage = 3
	case 3:
		// Add recipe message
		var recipe string
		if len(c.Payload.Resources) > 0 {
			for k, v := range c.Payload.Resources {
				rGetResourceByID, err := services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
					ID: k,
				})
				if err != nil {
					return err
				}

				recipe += rGetResourceByID.GetResource().GetName() + " x " + strconv.Itoa(int(v)) + "\n"
			}
		}

		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "crafting.confirm_choose_resources")+"\n\n "+recipe)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}
		// Aggiorno stato
		c.CurrentState.Stage = 4
	case 4:
		// Il player ha avviato il crafting, stampa il tempo di attesa
		// Aggiorna stato
		endTime := helpers.GetEndTime(0, 10, 0)
		var msg tgbotapi.MessageConfig
		msg = services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "crafting.wait", endTime.Format("15:04:05")),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		endTimeProto, err := ptypes.TimestampProto(endTime)
		if err != nil {
			panic(err)
		}

		c.CurrentState.FinishAt = endTimeProto
		c.CurrentState.ToNotify = true
		c.CurrentState.Stage = 5
		c.Breaker.ToMenu = true
	case 5:
		// crafting completato
		// Rimuovo risorse usate al player
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

		items, err := json.Marshal(c.Payload.Resources)
		if err != nil {
			return err
		}

		// Creo la richiesta di craft
		rCraft, err := services.NnSDK.Craft(helpers.NewContext(1), &pb.CraftRequest{
			CraftType: c.Payload.Item,
			Category:  c.Payload.Category,
			Items:     string(items),
			PlayerID:  c.Player.ID,
		})
		if err != nil {
			return err
		}

		var text string
		switch c.Payload.Item {
		case "armors":
			text = helpers.Trans(c.Player.Language.Slug, "crafting.craft_completed", rCraft.GetArmor().GetName())
		case "weapons":
			text = helpers.Trans(c.Player.Language.Slug, "crafting.craft_completed", rCraft.GetWeapon().GetName())
		}
		msg := services.NewMessage(c.Player.ChatID, text)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}
		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
