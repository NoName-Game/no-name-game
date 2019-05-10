package controllers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/app/provider"

	"bitbucket.org/no-name-game/no-name/app/commands"
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Crafting
func Crafting(update tgbotapi.Update) {
	//====================================
	// Init Func!
	//====================================
	type craftingPayload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}

	message := update.Message
	routeName := "route.crafting"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)
	var payload craftingPayload
	helpers.UnmarshalPayload(state.Payload, &payload)
	var addResourceFlag bool

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage", helpers.Player.Language.Slug)
	switch state.Stage {
	case 0:
		if helpers.InArray(message.Text, []string{
			helpers.Trans("armors", helpers.Player.Language.Slug),
			helpers.Trans("weapons", helpers.Player.Language.Slug),
		}) {
			state.Stage = 1
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	case 1:
		if helpers.InArray(message.Text, helpers.GetAllTranslatedSlugCategoriesByLocale(helpers.Player.Language.Slug)) {
			state.Stage = 2
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	case 2:
		if strings.Contains(message.Text, helpers.Trans("crafting.add", helpers.Player.Language.Slug)) {
			addResourceFlag = true
			validationFlag = true
		} else if message.Text == helpers.Trans("crafting.craft", helpers.Player.Language.Slug) {
			if len(payload.Resources) > 0 {
				state.Stage = 3
				state, _ = provider.UpdatePlayerState(state)
				validationFlag = true
			}
		}
	case 3:
		if message.Text == helpers.Trans("confirm", helpers.Player.Language.Slug) {
			state.FinishAt = commands.GetEndTime(0, 1, 10)
			state.Stage = 4

			// Stupid poninter stupid json pff
			t := new(bool)
			*t = true
			state.ToNotify = t

			state, _ = provider.UpdatePlayerState(state)
			validationMessage = helpers.Trans("crafting.wait", helpers.Player.Language.Slug, state.FinishAt.Format("15:04:05"))
			validationFlag = false
		}
	case 4:
		if time.Now().After(state.FinishAt) {
			validationFlag = true
		} else {
			validationMessage = helpers.Trans("crafting.wait", helpers.Player.Language.Slug, state.FinishAt.Format("15:04:05"))
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			validatorMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			services.SendMessage(validatorMsg)
		}
	}

	// Logic flux
	//		0		1		 	2			3
	// -> What -> Category -> Resources -> craft

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(craftingPayload{})
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.what", helpers.Player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("armors", helpers.Player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("weapons", helpers.Player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", helpers.Player.Language.Slug)),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", helpers.Player.Language.Slug)),
			),
		)
		services.SendMessage(msg)
	case 1:
		// If is valid input
		if validationFlag {
			payload.Item = message.Text
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = provider.UpdatePlayerState(state)
		}

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch payload.Item {
		case helpers.Trans("armors", helpers.Player.Language.Slug):
			armorCategories, err := provider.GetAllArmorCategory()
			if err != nil {
				services.ErrorHandler("Cant get armor categories", err)
			}

			for _, category := range armorCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(category.Slug, helpers.Player.Language.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons", helpers.Player.Language.Slug):
			weaponCategories, err := provider.GetAllWeaponCategory()
			if err != nil {
				services.ErrorHandler("Cant get weapon categories", err)
			}

			for _, category := range weaponCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(category.Slug, helpers.Player.Language.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", helpers.Player.Language.Slug)),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", helpers.Player.Language.Slug)),
		))

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.type", helpers.Player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)

	case 2:

		////////////////////////////////////
		// ONLY FOR DEBUG - Add one resource
		_, err := provider.AddResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
			ItemID:   42,
			Quantity: 2,
		})
		if err != nil {
			services.ErrorHandler("Cant add resource to player inventory", err)
		}
		////////////////////////////////////

		inventory, err := provider.GetPlayerInventory(helpers.Player)
		if err != nil {
			services.ErrorHandler("Cant get player inventory", err)
		}

		playerResources := helpers.InventoryToMap(inventory)

		// If is valid input
		if validationFlag {
			// Id Add new resource
			if addResourceFlag {
				if payload.Resources == nil {
					payload.Resources = make(map[uint]int)
				}

				// Clear text from Add and other shit.
				resourceName := strings.Split(
					strings.Split(message.Text, " (")[0],
					helpers.Trans("crafting.add", helpers.Player.Language.Slug)+" ")[1]

				resource, err := provider.FindResourceByName(resourceName)
				if err != nil {
					services.ErrorHandler("Cant find resource", err)
				}

				resourceID := resource.ID
				resourceMaxQuantity := playerResources[resourceID]

				if helpers.KeyInMap(resourceID, payload.Resources) {
					if payload.Resources[resourceID] < resourceMaxQuantity {
						payload.Resources[resourceID]++
					}
				} else {
					payload.Resources[resourceID] = 1
				}
			} else {
				payload.Category = helpers.Slugger(message.Text)
			}
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = provider.UpdatePlayerState(state)
		}

		// Keyboard with resources
		var keyboardRowResources [][]tgbotapi.KeyboardButton
		for r, q := range playerResources {
			// If PayloadResouces < Inventory quantity ok :)
			if payload.Resources[r] < q {
				resource, err := provider.GetResourceByID(r)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.add", helpers.Player.Language.Slug) + " " + resource.Name + " (" + (strconv.Itoa(q - payload.Resources[r])) + ")",
				))
				keyboardRowResources = append(keyboardRowResources, keyboardRow)
			}
		}

		// If PayloadResources is not empty show craft button
		if len(payload.Resources) > 0 {
			keyboardRowResources = append(keyboardRowResources, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.craft", helpers.Player.Language.Slug),
				),
			))
		}

		// Clear and exit
		keyboardRowResources = append(keyboardRowResources,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", helpers.Player.Language.Slug)),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", helpers.Player.Language.Slug)),
			),
		)

		//Add recipe message
		var recipe string
		if len(payload.Resources) > 0 {
			for k, v := range payload.Resources {
				resource, err := provider.GetResourceByID(k)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				recipe += resource.Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.choose_resources", helpers.Player.Language.Slug)+"\n"+recipe)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowResources,
		}
		services.SendMessage(msg)

	case 3:
		//Add recipe message
		var recipe string
		if len(payload.Resources) > 0 {
			for k, v := range payload.Resources {
				resource, err := provider.GetResourceByID(k)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				recipe += resource.Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.confirm_choose_resources", helpers.Player.Language.Slug)+"\n\n "+recipe)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("confirm", helpers.Player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", helpers.Player.Language.Slug)),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", helpers.Player.Language.Slug)),
			),
		)
		services.SendMessage(msg)
	case 4:
		if validationFlag {
			var craftingResult string

			switch payload.Item {
			case helpers.Trans("armors", helpers.Player.Language.Slug):

				var craftingRequest nnsdk.ArmorCraft
				helpers.UnmarshalPayload(state.Payload, &craftingRequest)
				crafted, err := provider.CraftArmor(craftingRequest)
				if err != nil {
					services.ErrorHandler("Cant create armor craft", err)
				}

				// Associate craft result tu player
				crafted.PlayerID = helpers.Player.ID
				crafted, err = provider.UpdateArmor(crafted)
				if err != nil {
					services.ErrorHandler("Cant associate armor craft", err)
				}

				// For message
				craftingResult = "Name: " + crafted.Name + "\nCategory: " + crafted.ArmorCategory.Name + "\nRarity: " + crafted.Rarity.Name
			case helpers.Trans("weapons", helpers.Player.Language.Slug):

				var craftingRequest nnsdk.WeaponCraft
				helpers.UnmarshalPayload(state.Payload, &craftingRequest)
				crafted, err := provider.CraftWeapon(craftingRequest)
				if err != nil {
					services.ErrorHandler("Cant create weapon craft", err)
				}

				// Associate craft result tu player
				crafted.PlayerID = helpers.Player.ID
				crafted, err = provider.UpdateWeapon(crafted)
				if err != nil {
					services.ErrorHandler("Cant associate armor craft", err)
				}

				// For message
				craftingResult = "Name: " + crafted.Name + "\nCategory: " + crafted.WeaponCategory.Name + "\nRarity: " + crafted.Rarity.Name
			}

			// Remove resources from player inventory
			for k, q := range payload.Resources {
				_, err := provider.RemoveResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
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

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.craft_completed", helpers.Player.Language.Slug)+"\n\n"+craftingResult)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", helpers.Player.Language.Slug)),
				),
			)
			services.SendMessage(msg)
		}
	}
}
