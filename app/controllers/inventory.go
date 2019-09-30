package controllers

import (
	"encoding/json"
	"strconv"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Inventory
func Inventory(update tgbotapi.Update) {
	message := update.Message

	msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.intro"))
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.inventory.recap")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.inventory.equip")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.inventory.destroy")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		),
	)

	services.SendMessage(msg)
}

// InventoryRecap - Send message with inventory recap
func InventoryRecap(update tgbotapi.Update) {
	message := update.Message

	var recap string

	// Summary Resources
	playerInventory, err := providers.GetPlayerInventory(helpers.Player)
	if err != nil {
		services.ErrorHandler("Can't get player inventory", err)
	}

	recap += "\n" + helpers.Trans("resources") + ":\n"
	playerResources := helpers.InventoryToMap(playerInventory)
	for r, q := range playerResources {
		resource, errResouce := providers.GetResourceByID(r)
		if errResouce != nil {
			services.ErrorHandler("Error in InventoryToString", err)
		}

		recap += "- " + resource.Name + " (" + (strconv.Itoa(q)) + ")\n"
	}

	// Summary Weapons
	playerWeapons, errWeapons := providers.GetPlayerWeapons(helpers.Player, "false")
	if errWeapons != nil {
		services.ErrorHandler("Can't get player weapons", err)
	}
	recap += "\n" + helpers.Trans("weapons") + ":\n"
	for _, weapon := range playerWeapons {
		recap += "- " + weapon.Name + "\n"
	}

	// Summary Armors
	playerArmors, errArmors := providers.GetPlayerArmors(helpers.Player, "false")
	if errArmors != nil {
		services.ErrorHandler("Can't get player armors", err)
	}

	recap += "\n" + helpers.Trans("armors") + ":\n"
	for _, armor := range playerArmors {
		recap += "- " + armor.Name + "\n"
	}

	msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.recap")+recap)
	services.SendMessage(msg)
}

// InventoryEquip - Menage player inventory
func InventoryEquip(update tgbotapi.Update) {
	//====================================
	// Init Func!
	//====================================
	type InventoryEquipPayload struct {
		Type    string
		EquipID uint
	}

	message := update.Message
	routeName := "route.inventory.equip"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)
	var payload InventoryEquipPayload
	helpers.UnmarshalPayload(state.Payload, &payload)

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
		if strings.Contains(message.Text, helpers.Trans("equip")) {
			state.Stage = 2
			state, _ = providers.UpdatePlayerState(state)
			validationFlag = true
		}
	case 2:
		if message.Text == helpers.Trans("confirm") {
			state.Stage = 3
			state, _ = providers.UpdatePlayerState(state)
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
	currentPlayerEquipment := helpers.Trans("inventory.equip.equipped")

	//////////////////////////////////
	currentPlayerEquipment += "\n" + helpers.Trans("armors") + ":\n"
	eqippedArmors, err := providers.GetPlayerArmors(helpers.Player, "true")
	if err != nil {
		services.ErrorHandler("Cant get equpped player armors", err)
	}

	for _, armor := range eqippedArmors {
		currentPlayerEquipment += "- " + armor.Name
	}
	//////////////////////////////////
	currentPlayerEquipment += "\n\n" + helpers.Trans("weapons") + ":\n"
	eqippedWeapons, err := providers.GetPlayerWeapons(helpers.Player, "true")
	if err != nil {
		services.ErrorHandler("Cant get equpped player weapons", err)
	}

	for _, weapon := range eqippedWeapons {
		currentPlayerEquipment += "- " + weapon.Name
	}
	//////////////////////////////////

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(InventoryEquipPayload{})
		state.Payload = string(payloadUpdated)
		state, _ = providers.UpdatePlayerState(state)

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.type")+currentPlayerEquipment)
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
			payload.Type = message.Text
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = providers.UpdatePlayerState(state)
		}

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch payload.Type {
		case helpers.Trans("armors"):
			armors, err := providers.GetPlayerArmors(helpers.Player, "false")
			if err != nil {
				services.ErrorHandler("Cant get player armors", err)
			}

			// Each player armors
			for _, armor := range armors {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("equip") + " " + armor.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons"):
			weapons, err := providers.GetPlayerWeapons(helpers.Player, "false")
			if err != nil {
				services.ErrorHandler("Cant get player weapons", err)
			}

			// Each player weapons
			for _, weapon := range weapons {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("equip") + " " + weapon.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		))

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.what"))
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
			equipmentName = strings.Split(message.Text, helpers.Trans("equip")+" ")[1]

			var equipmentID uint
			switch payload.Type {
			case helpers.Trans("armors"):
				var armor nnsdk.Armor
				armor, err := providers.FindArmorByName(equipmentName)
				if err != nil {
					services.ErrorHandler("Cant find equip armor name", err)
				}

				equipmentID = armor.ID
			case helpers.Trans("weapons"):
				var weapon nnsdk.Weapon
				weapon, err := providers.FindWeaponByName(equipmentName)
				if err != nil {
					services.ErrorHandler("Cant find equip weapon name", err)
				}

				equipmentID = weapon.ID
			}

			payload.EquipID = equipmentID
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = providers.UpdatePlayerState(state)
		}

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.equip.confirm")+"\n\n "+equipmentName)
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
	case 3:
		// If is valid input
		if validationFlag {
			switch payload.Type {
			case helpers.Trans("armors"):
				equipment, err := providers.GetArmorByID(payload.EquipID)
				if err != nil {
					services.ErrorHandler("Cant find armor by ID", err)
				}

				// Stupid poninter stupid json pff
				t := new(bool)
				*t = true
				equipment.Equipped = t

				_, err = providers.UpdateArmor(equipment)
				if err != nil {
					services.ErrorHandler("Cant update armor", err)
				}
			case helpers.Trans("weapons"):
				equipment, err := providers.GetWeaponByID(payload.EquipID)
				if err != nil {
					services.ErrorHandler("Cant find weapon by ID", err)
				}

				// Stupid poninter stupid json pff
				t := new(bool)
				*t = true
				equipment.Equipped = t

				_, err = providers.UpdateWeapon(equipment)
				if err != nil {
					services.ErrorHandler("Cant update weapon", err)
				}
			}
		}

		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(state, helpers.Player)
		//====================================

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.equip.completed"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)
	}
}

// InventoryDestroy - Destroy player item
func InventoryDestroy(update tgbotapi.Update) {
	//====================================
	// Init Func!
	//====================================
	type InventoryDestroyPayload struct {
		Type    string
		EquipID uint
	}

	message := update.Message
	routeName := "route.inventory.destroy"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)
	var payload InventoryDestroyPayload
	helpers.UnmarshalPayload(state.Payload, &payload)

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
		if strings.Contains(message.Text, helpers.Trans("destroy")) {
			state.Stage = 2
			state, _ = providers.UpdatePlayerState(state)
			validationFlag = true
		}
	case 2:
		if message.Text == helpers.Trans("confirm") {
			state.Stage = 3
			state, _ = providers.UpdatePlayerState(state)
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
		state, _ = providers.UpdatePlayerState(state)

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.destroy.type"))
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
			payload.Type = message.Text
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = providers.UpdatePlayerState(state)
		}

		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch payload.Type {
		case helpers.Trans("armors"):
			armors, err := providers.GetPlayerArmors(helpers.Player, "false")
			if err != nil {
				services.ErrorHandler("Cant get player armors", err)
			}

			// Each player armors
			for _, armor := range armors {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("destroy") + " " + armor.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		case helpers.Trans("weapons"):
			weapons, err := providers.GetPlayerWeapons(helpers.Player, "false")
			if err != nil {
				services.ErrorHandler("Cant get player weapons", err)
			}

			// Each player weapons
			for _, weapon := range weapons {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("destroy") + " " + weapon.Name))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		))

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.destroy.what"))
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
			equipmentName = strings.Split(message.Text, helpers.Trans("destroy")+" ")[1]

			var equipmentID uint
			switch payload.Type {
			case helpers.Trans("armors"):
				var armor nnsdk.Armor
				armor, err := providers.FindArmorByName(equipmentName)
				if err != nil {
					services.ErrorHandler("Cant find equip armor name", err)
				}

				equipmentID = armor.ID
			case helpers.Trans("weapons"):
				var weapon nnsdk.Weapon
				weapon, err := providers.FindWeaponByName(equipmentName)
				if err != nil {
					services.ErrorHandler("Cant find equip weapon name", err)
				}

				equipmentID = weapon.ID
			}

			payload.EquipID = equipmentID
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = providers.UpdatePlayerState(state)
		}

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.destroy.confirm")+"\n\n "+equipmentName)
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
	case 3:
		// If is valid input
		if validationFlag {
			switch payload.Type {
			case helpers.Trans("armors"):
				equipment, err := providers.GetArmorByID(payload.EquipID)
				if err != nil {
					services.ErrorHandler("Cant find weapon by ID", err)
				}

				_, err = providers.DeleteArmor(equipment)
				if err != nil {
					services.ErrorHandler("Cant delete armor", err)
				}
			case helpers.Trans("weapons"):
				equipment, err := providers.GetWeaponByID(payload.EquipID)
				if err != nil {
					services.ErrorHandler("Cant find weapon by ID", err)
				}

				_, err = providers.DeleteWeapon(equipment)
				if err != nil {
					services.ErrorHandler("Cant delete weapon", err)
				}
			}
		}

		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(state, helpers.Player)
		//====================================

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("inventory.destroy.completed"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)
	}
}
