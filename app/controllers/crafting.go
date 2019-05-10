package controllers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/no-name-game/no-name/app/commands"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Crafting
func Crafting(update tgbotapi.Update, player models.Player) {
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
	state := helpers.StartAndCreatePlayerState(routeName, player)
	var payload craftingPayload
	helpers.UnmarshalPayload(state.Payload, &payload)
	var addResourceFlag bool

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage", player.Language.Slug)
	switch state.Stage {
	case 0:
		if helpers.InArray(message.Text, []string{
			helpers.Trans("armors", player.Language.Slug),
			helpers.Trans("weapons", player.Language.Slug),
		}) {
			state.Stage = 1
			state.Update()
			validationFlag = true
		}
	case 1:
		if helpers.InArray(message.Text, helpers.GetAllTranslatedSlugCategoriesByLocale(player.Language.Slug)) {
			state.Stage = 2
			state.Update()
			validationFlag = true
		}
	case 2:
		if strings.Contains(message.Text, helpers.Trans("crafting.add", player.Language.Slug)) {
			addResourceFlag = true
			validationFlag = true
		} else if message.Text == helpers.Trans("crafting.craft", player.Language.Slug) {
			if len(payload.Resources) > 0 {
				state.Stage = 3
				state.Update()
				validationFlag = true
			}
		}
	case 3:
		if message.Text == helpers.Trans("confirm", player.Language.Slug) {
			state.FinishAt = commands.GetEndTime(0, 1, 10)
			state.Stage = 4
			state.ToNotify = true
			state.Update()
			validationMessage = helpers.Trans("crafting.wait", player.Language.Slug, state.FinishAt.Format("15:04:05"))
			validationFlag = false
		}
	case 4:
		if time.Now().After(state.FinishAt) {
			validationFlag = true
		} else {
			validationMessage = helpers.Trans("crafting.wait", player.Language.Slug, state.FinishAt.Format("15:04:05"))
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
		state.Update()

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.what", player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("armors", player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("weapons", player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", player.Language.Slug)),
			),
		)
		services.SendMessage(msg)
	case 1:
		// If is valid input
		if validationFlag {
			payload.Item = message.Text
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Update()
		}

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch payload.Item {
		case helpers.Trans("armors", player.Language.Slug):
			armorCategories := models.GetAllArmorCategories()
			for _, category := range armorCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(category.Slug, player.Language.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons", player.Language.Slug):
			weaponCategories := models.GetAllWeaponCategories()
			for _, category := range weaponCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(category.Slug, player.Language.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", player.Language.Slug)),
		))

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.type", player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)

	case 2:
		//ONLY FOR DEBUG - Add one resource
		player.Inventory.AddResource(models.GetResourceByID(42), 2)
		//player.Inventory.AddResource(models.GetResourceByID(46), 3)

		playerResources := player.Inventory.ToMap()

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
					helpers.Trans("crafting.add", player.Language.Slug)+" ")[1]

				resourceID := models.GetResourceByName(resourceName).ID
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
			state.Update()
		}

		// Keyboard with resources
		var keyboardRowResources [][]tgbotapi.KeyboardButton
		for r, q := range playerResources {
			// If PayloadResouces < Inventory quantity ok :)
			if payload.Resources[r] < q {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.add", player.Language.Slug) + " " + models.GetResourceByID(r).Name + " (" + (strconv.Itoa(q - payload.Resources[r])) + ")",
				))
				keyboardRowResources = append(keyboardRowResources, keyboardRow)
			}
		}

		// If PayloadResources is not empty show craft button
		if len(payload.Resources) > 0 {
			keyboardRowResources = append(keyboardRowResources, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.craft", player.Language.Slug),
				),
			))
		}

		// Clear and exit
		keyboardRowResources = append(keyboardRowResources,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", player.Language.Slug)),
			),
		)

		//Add recipe message
		var recipe string
		if len(payload.Resources) > 0 {
			for k, v := range payload.Resources {
				recipe += models.GetResourceByID(k).Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.choose_resources", player.Language.Slug)+"\n"+recipe)
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
				recipe += models.GetResourceByID(k).Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.confirm_choose_resources", player.Language.Slug)+"\n\n "+recipe)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("confirm", player.Language.Slug)),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", player.Language.Slug)),
			),
		)
		services.SendMessage(msg)
	case 4:
		if validationFlag {
			var craftingResult string

			switch payload.Item {
			case helpers.Trans("armors", player.Language.Slug):
				crafted := helpers.CraftArmor(state.Payload)

				// Associate craft result tu player
				crafted.AddPlayer(player)

				// For message
				craftingResult = "Name: " + crafted.Name + "\nCategory: " + crafted.ArmorCategory.Name + "\nRarity: " + crafted.Rarity.Name
			case helpers.Trans("weapons", player.Language.Slug):
				crafted := helpers.CraftWeapon(state.Payload)

				// Associate craft result tu player
				crafted.AddPlayer(player)

				// For message
				craftingResult = "Name: " + crafted.Name + "\nCategory: " + crafted.WeaponCategory.Name + "\nRarity: " + crafted.Rarity.Name
			}

			// Remove resources from player inventory
			for k, q := range payload.Resources {
				player.Inventory.RemoveItem(models.GetResourceByID(k), q)
			}

			//====================================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, player)
			//====================================

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("crafting.craft_completed", player.Language.Slug)+"\n\n"+craftingResult)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
				),
			)
			services.SendMessage(msg)
		}
	}
}
