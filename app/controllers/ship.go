package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"

	"bitbucket.org/no-name-game/no-name/app/commands"
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/provider"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ShipController struct{}

// Ship
func Ship(update tgbotapi.Update) {
	message := update.Message

	//====================================
	// Extra data
	//====================================
	currentShipRecap := "\n\n"
	eqippedShips, err := provider.GetPlayerShips(helpers.Player, true)
	if err != nil {
		services.ErrorHandler("Cant get equipped player ship", err)
	}

	for _, ship := range eqippedShips {
		currentShipRecap += helpers.Trans("name") + ": " + ship.Name + "\n"
		currentShipRecap += helpers.Trans("category") + ": " + ship.ShipCategory.Name + "\n"
		currentShipRecap += helpers.Trans("rarity") + ": " + ship.Rarity.Name + "\n"
		currentShipRecap += helpers.Trans("integrity") + ": " + strconv.FormatUint(uint64(ship.ShipStats.Integrity), 10) + "\n"
	}
	//////////////////////////////////

	msg := services.NewMessage(message.Chat.ID, helpers.Trans("ship.report")+currentShipRecap)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.ship.exploration")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.ship.warehouse")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.ship.repairs")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.ship.better")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		),
	)

	services.SendMessage(msg)
}

// ShipExploration
func ShipExploration(update tgbotapi.Update) {
	//====================================
	// Init Func!
	//====================================
	type craftingPayload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}

	message := update.Message
	routeName := "route.ship.exploration"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)
	var payload craftingPayload
	helpers.UnmarshalPayload(state.Payload, &payload)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage")
	switch state.Stage {
	case 0:
		if helpers.InArray(message.Text, []string{
			helpers.Trans("ship.exploration.start"),
		}) {
			state.Stage = 1
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	case 1:
		// if helpers.InArray(message.Text, helpers.GetAllTranslatedSlugCategoriesByLocale()) {
		// 	state.Stage = 2
		// 	state, _ = provider.UpdatePlayerState(state)
		// 	validationFlag = true
		// }
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
		payloadUpdated, _ := json.Marshal(craftingPayload{})
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)

		//====================================
		// Extra data
		//====================================
		currentPlayerPositions := "\n\n"
		position, err := provider.GetPlayerLastPosition(helpers.Player)
		if err != nil {
			services.ErrorHandler("Cant get player last position", err)
		}

		currentPlayerPositions += fmt.Sprintf("%s \nX: %v \nY: %v \nZ: %v \n", helpers.Trans("ship.exploration.current_position"), position.X, position.Y, position.Z)
		//////////////////////////////////

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("ship.exploration.info")+currentPlayerPositions)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("ship.exploration.start")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)
	case 1:
		if validationFlag {

			//====================================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, helpers.Player)
			//====================================

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("todo_text"))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
				),
			)
			services.SendMessage(msg)
		}
	}
}

// ShipWarehouse
func ShipWarehouse(update tgbotapi.Update) {
	//====================================
	// Init Func!
	//====================================
	type craftingPayload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}

	message := update.Message
	routeName := "route.ship.warehouse"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)
	var payload craftingPayload
	helpers.UnmarshalPayload(state.Payload, &payload)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage")
	switch state.Stage {
	case 0:
		if helpers.InArray(message.Text, []string{
			helpers.Trans("ship.exploration.start"),
		}) {
			state.Stage = 1
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	case 1:
		// if helpers.InArray(message.Text, helpers.GetAllTranslatedSlugCategoriesByLocale()) {
		// 	state.Stage = 2
		// 	state, _ = provider.UpdatePlayerState(state)
		// 	validationFlag = true
		// }
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
		payloadUpdated, _ := json.Marshal(craftingPayload{})
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("ship.warehouse.info"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)
	case 1:
		if validationFlag {

			//====================================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, helpers.Player)
			//====================================

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("todo_text"))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
				),
			)
			services.SendMessage(msg)
		}
	}
}

// ShipRepairs
func ShipRepairs(update tgbotapi.Update) {
	//====================================
	// Init Func!
	//====================================
	type repairsPayload struct {
		Ship              nnsdk.Ship
		QuantityResources float64
		RepairTime        float64
		TypeResources     string
	}

	// Stupid poninter stupid json pff
	t := new(bool)
	*t = true

	message := update.Message
	routeName := "route.ship.repairs"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)
	var payload repairsPayload
	helpers.UnmarshalPayload(state.Payload, &payload)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage")
	switch state.Stage {
	case 0:
		if helpers.InArray(message.Text, []string{
			helpers.Trans("ship.repairs.start"),
		}) {
			state.Stage = 1
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	case 1:
		if state.FinishAt.Before(time.Now()) {
			state.Stage = 2
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		} else {
			validationMessage = helpers.Trans("wait", state.FinishAt.Format("15:04:05"))
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
		//====================================
		// Extra data
		//====================================
		needRepair := true
		currentShipRecap := "\n\n"
		eqippedShips, err := provider.GetPlayerShips(helpers.Player, true)
		if err != nil {
			services.ErrorHandler("Cant get equipped player ship", err)
		}

		repairInfo, err := provider.GetShipRepairInfo(eqippedShips[0])
		if err != nil {
			services.ErrorHandler("Cant get ship repair info", err)
		}

		if repairInfo["QuantityResources"].(float64) <= 0 {
			needRepair = false
			currentShipRecap += helpers.Trans("ship.repairs.dont_need")
		} else {
			currentShipRecap += fmt.Sprintf("%s: %v\n", helpers.Trans("integrity"), eqippedShips[0].ShipStats.Integrity)
			currentShipRecap += fmt.Sprintf("%s: %v %s\n", helpers.Trans("ship.repairs.time"), repairInfo["RepairTime"], helpers.Trans("minutes"))
			currentShipRecap += fmt.Sprintf("%s: %v (%v)\n", helpers.Trans("ship.repairs.quantity_resources"), repairInfo["QuantityResources"], repairInfo["TypeResources"])
		}

		//////////////////////////////////

		payloadUpdated, _ := json.Marshal(repairsPayload{
			Ship:              eqippedShips[0],
			QuantityResources: repairInfo["QuantityResources"].(float64),
			RepairTime:        repairInfo["RepairTime"].(float64),
			TypeResources:     repairInfo["TypeResources"].(string),
		})

		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)

		var keyboardRow [][]tgbotapi.KeyboardButton
		if needRepair {
			newKeyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("ship.repairs.start")))
			keyboardRow = append(keyboardRow, newKeyboardRow)
		}

		// Clear and exit
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		))

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("ship.repairs.info")+currentShipRecap)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		services.SendMessage(msg)

	case 1:
		if validationFlag {
			//====================================
			// Extra data
			//====================================
			// START Repair ship
			responseStart, err := provider.StartShipRepair(payload.Ship)
			if err != nil {
				services.ErrorHandler("Cant repair ship", err)
			}

			recapResourceUsed := fmt.Sprintf("%s\n", helpers.Trans("ship.repairs.used_resources"))
			for resourceID, quantity := range responseStart {
				resource, err := provider.GetResourceByID(resourceID)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				recapResourceUsed += fmt.Sprintf("- %s : %v\n", resource.Name, quantity)
			}
			//////////////////////////////////

			// Set timer
			state.FinishAt = commands.GetEndTime(0, int(payload.RepairTime), 0)
			state.ToNotify = t
			state.Stage = 1

			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state, _ = provider.UpdatePlayerState(state)

			msg := services.NewMessage(message.Chat.ID,
				fmt.Sprintf("%s \n %s", recapResourceUsed, helpers.Trans("ship.repairs.reparing", state.FinishAt.Format("15:04:05"))),
			)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				),
			)
			services.SendMessage(msg)
		}
	case 2:
		if validationFlag {
			// END Repair ship
			_, err := provider.EndShipRepair(payload.Ship)
			if err != nil {
				services.ErrorHandler("Cant repair ship", err)
			}

			//====================================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, helpers.Player)
			//====================================

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("ship.repairs.reparing.finish"))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
				),
			)
			services.SendMessage(msg)
		}
	}
}

// ShipBetter
func ShipBetter(update tgbotapi.Update) {
	//====================================
	// Init Func!
	//====================================
	type craftingPayload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}

	message := update.Message
	routeName := "route.ship.better"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)
	var payload craftingPayload
	helpers.UnmarshalPayload(state.Payload, &payload)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage")
	switch state.Stage {
	case 0:
		if helpers.InArray(message.Text, []string{
			helpers.Trans("ship.exploration.start"),
		}) {
			state.Stage = 1
			state, _ = provider.UpdatePlayerState(state)
			validationFlag = true
		}
	case 1:
		// if helpers.InArray(message.Text, helpers.GetAllTranslatedSlugCategoriesByLocale()) {
		// 	state.Stage = 2
		// 	state, _ = provider.UpdatePlayerState(state)
		// 	validationFlag = true
		// }
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
		payloadUpdated, _ := json.Marshal(craftingPayload{})
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("ship.better.info"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)
	case 1:
		if validationFlag {

			//====================================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, helpers.Player)
			//====================================

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("todo_text"))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
				),
			)
			services.SendMessage(msg)
		}
	}
}
