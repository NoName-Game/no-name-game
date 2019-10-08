package controllers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type CraftingController struct {
	Update     tgbotapi.Update
	Message    *tgbotapi.Message
	RouteName  string
	Validation struct {
		HasErrors bool
		Message   string
	}
	Payload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}
	// Additional Data
	AddResourceFlag bool
}

//====================================
// Handle
//====================================
func (c *CraftingController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	c.RouteName = "route.crafting"
	c.Update = update
	c.Message = update.Message

	// Set Additional Data
	c.AddResourceFlag = false

	// Check current state for this routes
	state, isNewState := helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// It's first message
	if isNewState {
		c.Stage(state)
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(state.Payload, c.Payload)

	// Go to validator
	c.Validation.HasErrors, state = c.Validator(state)
	if !c.Validation.HasErrors {
		state, _ = providers.UpdatePlayerState(state)
		c.Stage(state)
		return
	}

	// Validator goes errors
	validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
	services.SendMessage(validatorMsg)
	return
}

//====================================
// Validator
//====================================
func (c *CraftingController) Validator(state nnsdk.PlayerState) (hasErrors bool, newState nnsdk.PlayerState) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch state.Stage {
	case 0:
		if helpers.InArray(c.Message.Text, []string{
			helpers.Trans("armors"),
			helpers.Trans("weapons"),
		}) {
			state.Stage = 1
			return false, state
		}
	case 1:
		if helpers.InArray(c.Message.Text, helpers.GetAllTranslatedSlugCategoriesByLocale()) {
			state.Stage = 2
			return false, state
		}
	case 2:
		if strings.Contains(c.Message.Text, helpers.Trans("crafting.add")) {
			c.AddResourceFlag = true
			return false, state
		} else if c.Message.Text == helpers.Trans("crafting.craft") && len(c.Payload.Resources) > 0 {
			state.Stage = 3
			return false, state
		}
	case 3:
		if c.Message.Text == helpers.Trans("confirm") {
			state.FinishAt = helpers.GetEndTime(0, 1, 10)
			state.Stage = 4

			// Stupid poninter stupid json pff
			t := new(bool)
			*t = true
			state.ToNotify = t

			// Force Update
			state, _ = providers.UpdatePlayerState(state)
			c.Validation.Message = helpers.Trans("crafting.wait", state.FinishAt.Format("15:04:05"))
			return true, state
		}
	case 4:
		c.Validation.Message = helpers.Trans("crafting.wait", state.FinishAt.Format("15:04:05"))
		if time.Now().After(state.FinishAt) {
			return false, state
		}
	}

	return true, state
}

//====================================
// Stage  0 -> 1 - What -> 2 - Category -> 3 - Resources -> 4 - Craft
//====================================
func (c *CraftingController) Stage(state nnsdk.PlayerState) {
	switch state.Stage {
	case 0:
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.what"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("armors")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("weapons")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)
	case 1:
		c.Payload.Item = c.Message.Text
		payloadUpdated, _ := json.Marshal(c.Payload)
		state.Payload = string(payloadUpdated)
		state, _ = providers.UpdatePlayerState(state)

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch c.Payload.Item {
		case helpers.Trans("armors"):
			armorCategories, err := providers.GetAllArmorCategory()
			if err != nil {
				services.ErrorHandler("Cant get armor categories", err)
			}

			for _, category := range armorCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(category.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons"):
			weaponCategories, err := providers.GetAllWeaponCategory()
			if err != nil {
				services.ErrorHandler("Cant get weapon categories", err)
			}

			for _, category := range weaponCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(category.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		))

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.type"))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)

	case 2:
		////////////////////////////////////
		// ONLY FOR DEBUG - Add one resource
		_, err := providers.AddResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
			ItemID:   42,
			Quantity: 2,
		})
		if err != nil {
			services.ErrorHandler("Cant add resource to player inventory", err)
		}
		////////////////////////////////////

		// Get Player inventory
		inventory, err := providers.GetPlayerInventory(helpers.Player)
		if err != nil {
			services.ErrorHandler("Cant get player inventory", err)
		}

		// Get player resource from inventory
		playerResources := helpers.InventoryToMap(inventory)

		// Add new resource
		if c.AddResourceFlag {
			if c.Payload.Resources == nil {
				c.Payload.Resources = make(map[uint]int)
			}

			// Clear text from Add and other shit.
			resourceName := strings.Split(
				strings.Split(c.Message.Text, " (")[0],
				helpers.Trans("crafting.add")+" ")[1]

			resource, err := providers.FindResourceByName(resourceName)
			if err != nil {
				services.ErrorHandler("Cant find resource", err)
			}

			resourceID := resource.ID
			resourceMaxQuantity := playerResources[resourceID]

			if helpers.KeyInMap(resourceID, c.Payload.Resources) {
				if c.Payload.Resources[resourceID] < resourceMaxQuantity {
					c.Payload.Resources[resourceID]++
				}
			} else {
				c.Payload.Resources[resourceID] = 1
			}
		} else {
			c.Payload.Category = helpers.Slugger(c.Message.Text)
		}

		payloadUpdated, _ := json.Marshal(c.Payload)
		state.Payload = string(payloadUpdated)
		state, _ = providers.UpdatePlayerState(state)

		// Keyboard with resources
		var keyboardRowResources [][]tgbotapi.KeyboardButton
		for r, q := range playerResources {
			// If PayloadResouces < Inventory quantity ok :)
			if c.Payload.Resources[r] < q {
				resource, err := providers.GetResourceByID(r)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.add") + " " + resource.Name + " (" + (strconv.Itoa(q - c.Payload.Resources[r])) + ")",
				))
				keyboardRowResources = append(keyboardRowResources, keyboardRow)
			}
		}

		// If PayloadResources is not empty show craft button
		// log.Panicln(c.Payload.Resources)
		if len(c.Payload.Resources) > 0 {
			keyboardRowResources = append(keyboardRowResources, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.craft"),
				),
			))
		}

		// Clear and exit
		keyboardRowResources = append(keyboardRowResources,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)

		//Add recipe message
		var recipe string
		if len(c.Payload.Resources) > 0 {
			for k, v := range c.Payload.Resources {
				resource, err := providers.GetResourceByID(k)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				recipe += resource.Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.choose_resources")+"\n"+recipe)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowResources,
		}
		services.SendMessage(msg)

	case 3:
		//Add recipe message
		var recipe string
		if len(c.Payload.Resources) > 0 {
			for k, v := range c.Payload.Resources {
				resource, err := providers.GetResourceByID(k)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				recipe += resource.Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.confirm_choose_resources")+"\n\n "+recipe)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)
	case 4:
		var craftingResult string

		switch c.Payload.Item {
		case helpers.Trans("armors"):
			// Call WS to craft armor
			var craftingRequest nnsdk.ArmorCraft
			helpers.UnmarshalPayload(state.Payload, &craftingRequest)
			crafted, err := providers.CraftArmor(craftingRequest)
			if err != nil {
				services.ErrorHandler("Cant create armor craft", err)
			}

			// Associate craft result tu player
			crafted.PlayerID = helpers.Player.ID
			crafted, err = providers.UpdateArmor(crafted)
			if err != nil {
				services.ErrorHandler("Cant associate armor craft", err)
			}

			// For message
			craftingResult = "Name: " + crafted.Name + "\nCategory: " + crafted.ArmorCategory.Name + "\nRarity: " + crafted.Rarity.Name
		case helpers.Trans("weapons"):
			// Call WS to craft weapon
			var craftingRequest nnsdk.WeaponCraft
			helpers.UnmarshalPayload(state.Payload, &craftingRequest)
			crafted, err := providers.CraftWeapon(craftingRequest)
			if err != nil {
				services.ErrorHandler("Cant create weapon craft", err)
			}

			// Associate craft result tu player
			crafted.PlayerID = helpers.Player.ID
			crafted, err = providers.UpdateWeapon(crafted)
			if err != nil {
				services.ErrorHandler("Cant associate armor craft", err)
			}

			// For message
			craftingResult = "Name: " + crafted.Name + "\nCategory: " + crafted.WeaponCategory.Name + "\nRarity: " + crafted.Rarity.Name
		}

		// Remove resources from player inventory
		for k, q := range c.Payload.Resources {
			_, err := providers.RemoveResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
				ItemID:   k,
				Quantity: q,
			})

			if err != nil {
				services.ErrorHandler("Cant add resource to player inventory", err)
			}
		}

		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(state, helpers.Player)
		//====================================

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.craft_completed")+"\n\n"+craftingResult)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)
	}
}
