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
	validationMessage := helpers.Trans("validationMessage")
	switch state.Stage {
	case 0:
		if helpers.InArray(message.Text, []string{
			helpers.Trans("armors"),
			helpers.Trans("weapons"),
		}) {
			state.Stage = 1
			state, _ = providers.UpdatePlayerState(state)
			validationFlag = true
		}
	case 1:
		if helpers.InArray(message.Text, helpers.GetAllTranslatedSlugCategoriesByLocale()) {
			state.Stage = 2
			state, _ = providers.UpdatePlayerState(state)
			validationFlag = true
		}
	case 2:
		if strings.Contains(message.Text, helpers.Trans("crafting.add")) {
			addResourceFlag = true
			validationFlag = true
		} else if message.Text == helpers.Trans("crafting.craft") {
			if len(payload.Resources) > 0 {
				state.Stage = 3
				state, _ = providers.UpdatePlayerState(state)
				validationFlag = true
			}
		}
	case 3:
		if message.Text == helpers.Trans("confirm") {
			state.FinishAt = helpers.GetEndTime(0, 1, 10)
			state.Stage = 4

			// Stupid poninter stupid json pff
			t := new(bool)
			*t = true
			state.ToNotify = t

			state, _ = providers.UpdatePlayerState(state)
			validationMessage = helpers.Trans("crafting.wait", state.FinishAt.Format("15:04:05"))
			validationFlag = false
		}
	case 4:
		if time.Now().After(state.FinishAt) {
			validationFlag = true
		} else {
			validationMessage = helpers.Trans("crafting.wait", state.FinishAt.Format("15:04:05"))
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
		state, _ = providers.UpdatePlayerState(state)

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.what"))
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
		// If is valid input
		if validationFlag {
			payload.Item = message.Text
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = providers.UpdatePlayerState(state)
		}

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch payload.Item {
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

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.type"))
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

		inventory, err := providers.GetPlayerInventory(helpers.Player)
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
					helpers.Trans("crafting.add")+" ")[1]

				resource, err := providers.FindResourceByName(resourceName)
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
			state, _ = providers.UpdatePlayerState(state)
		}

		// Keyboard with resources
		var keyboardRowResources [][]tgbotapi.KeyboardButton
		for r, q := range playerResources {
			// If PayloadResouces < Inventory quantity ok :)
			if payload.Resources[r] < q {
				resource, err := providers.GetResourceByID(r)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.add") + " " + resource.Name + " (" + (strconv.Itoa(q - payload.Resources[r])) + ")",
				))
				keyboardRowResources = append(keyboardRowResources, keyboardRow)
			}
		}

		// If PayloadResources is not empty show craft button
		if len(payload.Resources) > 0 {
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
		if len(payload.Resources) > 0 {
			for k, v := range payload.Resources {
				resource, err := providers.GetResourceByID(k)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				recipe += resource.Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.choose_resources")+"\n"+recipe)
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
				resource, err := providers.GetResourceByID(k)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				recipe += resource.Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.confirm_choose_resources")+"\n\n "+recipe)
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
		if validationFlag {
			var craftingResult string

			switch payload.Item {
			case helpers.Trans("armors"):

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
			for k, q := range payload.Resources {
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

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.craft_completed")+"\n\n"+craftingResult)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				),
			)
			services.SendMessage(msg)
		}
	}
}
