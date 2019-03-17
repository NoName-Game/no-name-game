package controllers

import (
	"encoding/json"
	"strconv"
	"strings"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-telegram-bot-api/telegram-bot-api"
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
	routeName := "crafting"
	state := helpers.StartAndCreatePlayerState(routeName, player)
	var payload craftingPayload
	helpers.UnmarshalPayload(state.Payload, &payload)
	var addResourceFlag bool

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := "Wrong input, please repeat or exit."
	switch state.Stage {
	case 0:
		if helpers.InArray(message.Text, []string{"Armors", "Weapons"}) {
			state.Stage = 1
			state.Update()
			validationFlag = true
		}
	case 1:
		if helpers.InArray(message.Text, helpers.GetAllCategories()) {
			state.Stage = 2
			state.Update()
			validationFlag = true
		}
	case 2:
		if strings.Contains(message.Text, "Add") {
			addResourceFlag = true
			validationFlag = true
		} else if message.Text == "Craft" {
			if len(payload.Resources) > 0 {
				state.Stage = 3
				state.Update()
				validationFlag = true
			}
		}
	case 3:
		if message.Text == "YES!" {
			state.Stage = 4
			state.Update()
			validationFlag = true
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
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

		msg := services.NewMessage(message.Chat.ID, "What do you want craft?")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Armors"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Weapons"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
				tgbotapi.NewKeyboardButton("clears"),
			),
		)
		services.SendMessage(msg)
	case 1:
		// If is valid input
		if validationFlag {
			payload.Item = helpers.Slugger(message.Text)
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Update()
		}

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch payload.Item {
		case "armors":
			armorCategories := models.GetAllArmorCategories()
			for _, category := range armorCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(category.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case "weapons":
			weaponCategories := models.GetAllWeaponCategories()
			for _, category := range weaponCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(category.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("back"),
			tgbotapi.NewKeyboardButton("clears"),
		))

		msg := services.NewMessage(message.Chat.ID, "Which type?")
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)

	case 2:
		//ONLY FOR DEBUG - Add one resource
		// player.Inventory.AddResource(models.GetResourceByID(42), 2)
		// player.Inventory.AddResource(models.GetResourceByID(46), 3)

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
					"Add ")[1]

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
					"Add " + models.GetResourceByID(r).Name + " (" + (strconv.Itoa(q - payload.Resources[r])) + ")",
				))
				keyboardRowResources = append(keyboardRowResources, keyboardRow)
			}
		}

		// If PayloadResources is not empty show craft button
		if len(payload.Resources) > 0 {
			keyboardRowResources = append(keyboardRowResources, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Craft"),
			))
		}

		// Clear and exit
		keyboardRowResources = append(keyboardRowResources,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
				tgbotapi.NewKeyboardButton("clears"),
			),
		)

		//Add recipe message
		var recipe string
		if len(payload.Resources) > 0 {
			for k, v := range payload.Resources {
				recipe += models.GetResourceByID(k).Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		msg := services.NewMessage(message.Chat.ID, "Choose which resources to use \n"+recipe)
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

		msg := services.NewMessage(message.Chat.ID, "Are you sure? You want use this recipe? \n\n "+recipe)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("YES!"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
				tgbotapi.NewKeyboardButton("clears"),
			),
		)
		services.SendMessage(msg)
	case 4:
		var craftingResult string

		switch payload.Item {
		case "armors":
			crafted := helpers.CraftArmor(state.Payload)

			// Associate craft result tu player
			crafted.AddPlayer(player)

			// For message
			craftingResult = "Name: " + crafted.Name + "\nCategory: " + crafted.ArmorCategory.Name + "\nRarity: " + crafted.Rarity.Name
		case "weapons":
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

		msg := services.NewMessage(message.Chat.ID, "Completed! This is your craft: \n\n"+craftingResult)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("back"),
			),
		)
		services.SendMessage(msg)
	}
}
