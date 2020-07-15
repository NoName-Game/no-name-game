package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
)

// ====================================
// CrafterController (NPC Crafter)
// ====================================
type CrafterController struct {
	Payload struct {
		Item      string
		Category  string
		Resources map[uint]int
		AddFlag   bool
	}
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *CrafterController) Handle(player nnsdk.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error
	var PlayerStateProvider providers.PlayerStateProvider

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.safeplanet.crafter",
		c.Payload,
		[]string{},
		c.Player,
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
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

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
	c.State.Payload = string(payloadUpdated)
	c.State, err = PlayerStateProvider.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

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

	switch c.State.Stage {
	case 1:
		if helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "armors"),
			helpers.Trans(c.Player.Language.Slug, "weapons"),
		}) {
			c.Payload.Item = c.Update.Message.Text
			return false, err
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true, err
	case 2:
		if helpers.InArray(c.Update.Message.Text, helpers.GetAllTranslatedSlugCategoriesByLocale(c.Player.Language.Slug)) {
			return false, err
		}
		return true, err
	case 3:
		if strings.Contains(c.Update.Message.Text, helpers.Trans(c.Player.Language.Slug, "crafting.add")) {
			c.State.Stage = 2
			c.Payload.AddFlag = true
			return false, err
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "crafting.craft") {
			if len(c.Payload.Resources) > 0 {
				return false, err
			}
		}
	case 4:
		if c.Update.Message.Text == helpers.Trans("confirm", c.Player.Language.Slug) {
			c.State.FinishAt = helpers.GetEndTime(0, 1, 10)
			*c.State.ToNotify = true
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "crafting.wait", c.Player.Language.Slug, c.State.FinishAt.Format("15:04:05"))
			return true, err
		}
	case 5:

	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *CrafterController) Stage() (err error) {
	switch c.State.Stage {
	// Invio messaggio con recap stats
	case 0:

		msg := services.NewMessage(c.Player.ChatID, helpers.Trans("crafting.what", c.Player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("armors", c.Player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("weapons", c.Player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", c.Player.Language.Slug)),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", c.Player.Language.Slug)),
			),
		)
		services.SendMessage(msg)
		// Avanzo di stage
		c.State.Stage = 1
	case 1:
		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch c.Payload.Item {
		case helpers.Trans("armors", c.Player.Language.Slug):
			var armorProvider providers.ArmorCategoryProvider
			armorCategories, _ := armorProvider.GetAllArmorCategory()
			for _, category := range armorCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(category.Slug, c.Player.Language.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons", c.Player.Language.Slug):
			var weaponProvider providers.ArmorCategoryProvider
			weaponCategories, _ := weaponProvider.GetAllArmorCategory()
			for _, category := range weaponCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(category.Slug, c.Player.Language.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", c.Player.Language.Slug)),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", c.Player.Language.Slug)),
		))

		msg := services.NewMessage(c.Player.ChatID, helpers.Trans("crafting.type", c.Player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)
		// Aggiorno stato
		c.State.Stage = 2
	case 2:
		var playerProvider providers.PlayerProvider
		playerResources, _ := playerProvider.GetPlayerResources(c.Player.ID)
		mapInventory := helpers.InventoryResourcesToMap(playerResources)
		var resourceProvider providers.ResourceProvider
		// Id Add new resource
		if c.Payload.AddFlag {
			if c.Payload.Resources == nil {
				c.Payload.Resources = make(map[uint]int)
			}

			// Clear text from Add and other shit.
			resourceName := strings.Split(
				strings.Split(c.Update.Message.Text, " (")[0],
				helpers.Trans("crafting.add", c.Player.Language.Slug)+" ")[1]

			resource, _ := resourceProvider.GetResourceByName(resourceName)
			resourceID := resource.ID
			resourceMaxQuantity := mapInventory[resourceID]

			if helpers.KeyInMap(resourceID, c.Payload.Resources) {
				if c.Payload.Resources[resourceID] < resourceMaxQuantity {
					c.Payload.Resources[resourceID]++
				}
			} else {
				c.Payload.Resources[resourceID] = 1
			}
		} else {
			c.Payload.Category = helpers.Slugger(c.Update.Message.Text)
		}

		// Keyboard with resources
		var keyboardRowResources [][]tgbotapi.KeyboardButton
		for r, q := range playerResources {
			// If PayloadResouces < Inventory quantity ok :)
			if c.Payload.Resources[uint(r)] < *q.Quantity {
				var resourceProvider providers.ResourceProvider
				resource, _ := resourceProvider.GetResourceByID(uint(r))
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.add", c.Player.Language.Slug) + " " + resource.Name + " (" + (strconv.Itoa(*q.Quantity - c.Payload.Resources[uint(r)])) + ")",
				))
				keyboardRowResources = append(keyboardRowResources, keyboardRow)
			}
		}

		// If PayloadResources is not empty show craft button
		if len(c.Payload.Resources) > 0 {
			keyboardRowResources = append(keyboardRowResources, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.craft", c.Player.Language.Slug),
				),
			))
		}

		// Clear and exit
		keyboardRowResources = append(keyboardRowResources,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", c.Player.Language.Slug)),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", c.Player.Language.Slug)),
			),
		)

		//Add recipe message
		var recipe string
		if len(c.Payload.Resources) > 0 {
			for k, v := range c.Payload.Resources {
				var resourceProvider providers.ResourceProvider
				resource, _ := resourceProvider.GetResourceByID(k)
				recipe += resource.Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(c.Player.ChatID, helpers.Trans("crafting.choose_resources", c.Player.Language.Slug)+"\n"+recipe)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowResources,
		}
		services.SendMessage(msg)
		// Aggiorno stato
		c.State.Stage = 3
	case 3:
		//Add recipe message
		var recipe string
		var resourceProvider providers.ResourceProvider
		if len(c.Payload.Resources) > 0 {
			for k, v := range c.Payload.Resources {
				resource, _ := resourceProvider.GetResourceByID(k)
				recipe += resource.Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(c.Player.ChatID, helpers.Trans("crafting.confirm_choose_resources", c.Player.Language.Slug)+"\n\n "+recipe)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("confirm", c.Player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", c.Player.Language.Slug)),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", c.Player.Language.Slug)),
			),
		)
		services.SendMessage(msg)
		// Aggiorno stato
		c.State.Stage = 4
	}

	return
}
