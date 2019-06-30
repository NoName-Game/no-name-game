package controllers

import (
	"encoding/json"
	"strconv"

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
	eqippedShips, err := provider.GetPlayerShips(helpers.Player, "true")
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

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("ship.exploration.info"))
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
	type craftingPayload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}

	message := update.Message
	routeName := "route.ship.repairs"
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

		msg := services.NewMessage(message.Chat.ID, helpers.Trans("ship.repairs.info"))
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
