package controllers

import (
	"encoding/json"
	"strconv"
	"strings"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/app/provider"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Inventory
func Inventory(update tgbotapi.Update, player nnsdk.Player) {
	message := update.Message

	msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.intro", player.Language.Slug))
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.inventory.recap", player.Language.Slug)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.inventory.equip", player.Language.Slug)),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.inventory.destroy", player.Language.Slug)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", player.Language.Slug)),
		),
	)

	services.SendMessage(msg)
}

// InventoryRecap - Send message with inventory recap
func InventoryRecap(update tgbotapi.Update, player nnsdk.Player) {
	message := update.Message

	var recap string

	// Summary Resources
	recap += "\n" + helpers.Trans("resources", player.Language.Slug) + ":\n"
	playerResources := helpers.InventoryToMap(player.Inventory)
	for r, q := range playerResources {

		resource, err := provider.GetResouceByID(r)
		if err != nil {
			services.ErrorHandler("Error in InventoryToString", err)
		}

		recap += "- " + resource.Name + " (" + (strconv.Itoa(q)) + ")\n"
	}

	// Summary Weapons
	recap += "\n" + helpers.Trans("weapons", player.Language.Slug) + ":\n"
	for _, weapon := range player.Weapons {
		recap += "- " + weapon.Name + "\n"
	}

	// Summary Armors
	recap += "\n" + helpers.Trans("armors", player.Language.Slug) + ":\n"
	for _, armor := range player.Armors {
		recap += "- " + armor.Name + "\n"
	}

	msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.recap", player.Language.Slug)+recap)
	services.SendMessage(msg)
}

// InventoryEquip - Menage player inventory
func InventoryEquip(update tgbotapi.Update, player nnsdk.Player) {
	//====================================
	// Init Func!
	//====================================
	type InventoryEquipPayload struct {
		Type    string
		EquipID uint
	}

	message := update.Message
	routeName := "route.inventory.equip"
	state := helpers.StartAndCreatePlayerState(routeName, player)
	var payload InventoryEquipPayload
	helpers.UnmarshalPayload(state.Payload, &payload)

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
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	case 1:
		if strings.Contains(message.Text, helpers.Trans("equip", player.Language.Slug)) {
			state.Stage = 2
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	case 2:
		if message.Text == helpers.Trans("confirm", player.Language.Slug) {
			state.Stage = 3
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			validatorMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			services.SendMessage(validatorMsg)
		}
	}

	//====================================
	// Extra data
	//====================================
	currentPlayerEquipment := helpers.Trans("inventory.equip.equipped", player.Language.Slug)

	currentPlayerEquipment += "\n" + helpers.Trans("armors", player.Language.Slug) + ":\n"

	//TODO: TO FIX
	// for _, armor := range player.GetEquippedArmors() {
	// 	currentPlayerEquipment += "- " + armor.Name
	// }

	// currentPlayerEquipment += "\n\n" + helpers.Trans("weapons", player.Language.Slug) + ":\n"
	// for _, weapon := range player.GetEquippedWeapons() {
	// 	currentPlayerEquipment += "- " + weapon.Name
	// }

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(InventoryEquipPayload{})
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.type", player.Language.Slug)+currentPlayerEquipment)
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
			payload.Type = message.Text
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = provider.UpdatePlayerState(state)
		}

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch payload.Type {
		case helpers.Trans("armors", player.Language.Slug):
			// Each player armors
			for _, armor := range player.Armors {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("equip", player.Language.Slug) + " " + armor.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons", player.Language.Slug):
			// Each player weapons
			for _, weapon := range player.Weapons {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("equip", player.Language.Slug) + " " + weapon.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", player.Language.Slug)),
		))

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.what", player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)
	case 2:
		var equipmentName string
		// If is valid input
		if validationFlag {
			// Clear text from Add and other shit.
			equipmentName = strings.Split(message.Text, helpers.Trans("equip", player.Language.Slug)+" ")[1]

			var equipmentID uint
			switch payload.Type {
			case helpers.Trans("armors", player.Language.Slug):
				equipmentID = models.GetArmorByName(equipmentName).ID
			case helpers.Trans("weapons", player.Language.Slug):
				equipmentID = models.GetWeaponByName(equipmentName).ID
			}

			payload.EquipID = equipmentID
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = provider.UpdatePlayerState(state)
		}

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.equip.confirm", player.Language.Slug)+"\n\n "+equipmentName)
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
	case 3:
		// If is valid input
		if validationFlag {
			switch payload.Type {
			case helpers.Trans("armors", player.Language.Slug):
				equipment := models.GetArmorByID(payload.EquipID)
				equipment.Equipped = true
				equipment.Update()
			case helpers.Trans("weapons", player.Language.Slug):
				equipment := models.GetWeaponByID(payload.EquipID)
				equipment.Equipped = true
				equipment.Update()
			}
		}

		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(state, player)
		//====================================

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.equip.completed", player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
			),
		)
		services.SendMessage(msg)
	}
}

// InventoryDestroy - Destroy player item
func InventoryDestroy(update tgbotapi.Update, player nnsdk.Player) {
	//====================================
	// Init Func!
	//====================================
	type InventoryDestroyPayload struct {
		Type    string
		EquipID uint
	}

	message := update.Message
	routeName := "route.inventory.destroy"
	state := helpers.StartAndCreatePlayerState(routeName, player)
	var payload InventoryDestroyPayload
	helpers.UnmarshalPayload(state.Payload, &payload)

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
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	case 1:
		if strings.Contains(message.Text, helpers.Trans("destroy", player.Language.Slug)) {
			state.Stage = 2
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	case 2:
		if message.Text == helpers.Trans("confirm", player.Language.Slug) {
			state.Stage = 3
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			validatorMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			services.SendMessage(validatorMsg)
		}
	}

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(InventoryDestroyPayload{})
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.destroy.type", player.Language.Slug))
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
			payload.Type = message.Text
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = provider.UpdatePlayerState(state)
		}

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch payload.Type {
		case helpers.Trans("armors", player.Language.Slug):
			// Each player armors
			for _, armor := range player.Armors {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("destroy", player.Language.Slug) + " " + armor.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons", player.Language.Slug):
			// Each player weapons
			for _, weapon := range player.Weapons {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("destroy", player.Language.Slug) + " " + weapon.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears", player.Language.Slug)),
		))

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.destroy.what", player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)
	case 2:
		var equipmentName string
		// If is valid input
		if validationFlag {
			// Clear text from Add and other shit.
			equipmentName = strings.Split(message.Text, helpers.Trans("destroy", player.Language.Slug)+" ")[1]

			var equipmentID uint
			switch payload.Type {
			case helpers.Trans("armors", player.Language.Slug):
				equipmentID = models.GetArmorByName(equipmentName).ID
			case helpers.Trans("weapons", player.Language.Slug):
				equipmentID = models.GetWeaponByName(equipmentName).ID
			}

			payload.EquipID = equipmentID
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = provider.UpdatePlayerState(state)
		}

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.destroy.confirm", player.Language.Slug)+"\n\n "+equipmentName)
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
	case 3:
		// If is valid input
		if validationFlag {
			switch payload.Type {
			case helpers.Trans("armors", player.Language.Slug):
				equipment := models.GetArmorByID(payload.EquipID)
				equipment.Delete()
			case helpers.Trans("weapons", player.Language.Slug):
				equipment := models.GetWeaponByID(payload.EquipID)
				equipment.Delete()
			}
		}

		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(state, player)
		//====================================

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.destroy.completed", player.Language.Slug))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
			),
		)
		services.SendMessage(msg)
	}
}
