package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type ShipController BaseController

//====================================
// Handle
//====================================
func (c *ShipController) Handle(update tgbotapi.Update) {
	message := update.Message

	//====================================
	// Extra data
	//====================================
	currentShipRecap := "\n\n"
	eqippedShips, err := providers.GetPlayerShips(helpers.Player, true)
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
			tgbotapi.NewKeyboardButton(helpers.Trans("route.ship.repairs")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		),
	)

	services.SendMessage(msg)
}

type ShipExplorationController struct {
	BaseController
	Payload struct {
		Ship               nnsdk.Ship
		StarNearestMapName map[int]string
		StarNearestMapInfo map[int]providers.ResponseExplorationInfo
		StarIDChosen       int
	}
}

//====================================
// Handle
//====================================
func (c *ShipExplorationController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	c.RouteName = "route.ship.exploration"
	c.Update = update
	c.Message = update.Message

	// Check current state for this routes
	state, isNewState := helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// Set and load payload
	helpers.UnmarshalPayload(state.Payload, &c.Payload)

	// It's first message
	if isNewState {
		c.Stage(state)
		return
	}

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
func (c *ShipExplorationController) Validator(state nnsdk.PlayerState) (hasErrors bool, newState nnsdk.PlayerState) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch state.Stage {
	case 0:
		if helpers.InArray(c.Message.Text, []string{
			helpers.Trans("ship.exploration.start"),
		}) {
			state.Stage = 1
			return false, state
		}
	case 1:
		if helpers.InArray(c.Message.Text, c.Payload.StarNearestMapName) {
			state.Stage = 2
			return false, state
		}
	case 2:
		c.Validation.Message = helpers.Trans("wait", state.FinishAt.Format("15:04:05"))
		if time.Now().After(state.FinishAt) {
			state.Stage = 3
			return false, state
		}
	}

	return true, state
}

//====================================
// Stage
//====================================
func (c *ShipExplorationController) Stage(state nnsdk.PlayerState) {
	switch state.Stage {
	case 0:
		currentPlayerPositions := "\n\n"
		position, err := providers.GetPlayerLastPosition(helpers.Player)
		if err != nil {
			services.ErrorHandler("Cant get player last position", err)
		}

		currentPlayerPositions += fmt.Sprintf("%s \nX: %v \nY: %v \nZ: %v \n", helpers.Trans("ship.exploration.current_position"), position.X, position.Y, position.Z)

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.exploration.info")+currentPlayerPositions)
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

		//====================================
		// Extra data
		//====================================
		eqippedShips, err := providers.GetPlayerShips(helpers.Player, true)
		if err != nil {
			services.ErrorHandler("Cant get equipped player ship", err)
		}

		msgNearestStars := "\n\n"
		explorationInfos, err := providers.GetShipExplorationInfo(eqippedShips[0])
		if err != nil {
			services.ErrorHandler("Cant get player last position", err)
		}

		// It's for match with keyboard in validator and needed for next step
		var starNearestMapName = make(map[int]string)
		var starNearestMapInfo = make(map[int]providers.ResponseExplorationInfo)

		// Keyboard with resources
		var keyboardRowStars [][]tgbotapi.KeyboardButton

		for _, explorationInfo := range explorationInfos {
			msgNearestStars += fmt.Sprintf("%s:%s\n", helpers.Trans("name"), explorationInfo.Star.Name)
			msgNearestStars += fmt.Sprintf("%s:%v\n", helpers.Trans("ship.exploration.fuel_needed"), explorationInfo.Fuel)
			msgNearestStars += fmt.Sprintf("%s:%v\n", helpers.Trans("ship.exploration.time_needed"), explorationInfo.Time)
			msgNearestStars += fmt.Sprintf("X: %v \nY: %v \nZ: %v \n\n", explorationInfo.Star.X, explorationInfo.Star.Y, explorationInfo.Star.Z)

			// Add for validation and next step
			starNearestMapName[int(explorationInfo.Star.ID)] = explorationInfo.Star.Name
			starNearestMapInfo[int(explorationInfo.Star.ID)] = explorationInfo

			// Add for keyboard
			keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
				explorationInfo.Star.Name,
			))
			keyboardRowStars = append(keyboardRowStars, keyboardRow)
		}
		//////////////////////////////////

		// Clear and exit
		keyboardRowStars = append(keyboardRowStars,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)

		// Update state
		c.Payload.Ship = eqippedShips[0]
		c.Payload.StarNearestMapName = starNearestMapName
		c.Payload.StarNearestMapInfo = starNearestMapInfo
		payloadUpdated, _ := json.Marshal(c.Payload)
		state.Payload = string(payloadUpdated)
		state, _ = providers.UpdatePlayerState(state)

		// Send message
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.exploration.research")+msgNearestStars)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowStars,
		}
		services.SendMessage(msg)

	case 2:
		// Filter chosen star id by message
		var starIDchosen int
		for key, name := range c.Payload.StarNearestMapName {
			if name == c.Message.Text {
				starIDchosen = key
				break
			}
		}

		// Not found
		if starIDchosen <= 0 {
			services.ErrorHandler("Cant get chose star destination", errors.New("Cant get chose star destination"))
		}

		// Set timer
		c.State.FinishAt = helpers.GetEndTime(0, int(c.Payload.StarNearestMapInfo[starIDchosen].Time), 0)
		t := new(bool)
		*t = true
		c.State.ToNotify = t
		c.State.Stage = 2

		c.Payload.StarIDChosen = starIDchosen

		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, _ = providers.UpdatePlayerState(c.State)

		msg := services.NewMessage(c.Message.Chat.ID,
			fmt.Sprintf("%s \n", helpers.Trans("ship.exploration.exploring", state.FinishAt.Format("15:04:05"))),
		)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)

	case 3:
		// End exploration
		var request providers.RequestExplorationEnd
		request.Position = []float64{
			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Star.X,
			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Star.Y,
			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Star.Z,
		}
		request.Tank = c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Fuel

		_, err := providers.EndShipExploration(c.Payload.Ship, request)
		if err != nil {
			services.ErrorHandler("Cant end exploration", err)
		}

		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(c.State, helpers.Player)
		//====================================

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.exploration.end"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)
	}
}

type ShipRepairsController struct {
	BaseController
	Payload struct {
		Ship              nnsdk.Ship
		QuantityResources float64
		RepairTime        float64
		TypeResources     string
	}
}

//====================================
// Handle
//====================================
func (c *ShipRepairsController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	c.RouteName = "route.ship.repairs"
	c.Update = update
	c.Message = update.Message

	// Check current state for this routes
	var isNewState bool
	c.State, isNewState = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// It's first message
	if isNewState {
		c.Stage(c.State)
		return
	}

	// Go to validator
	c.Validation.HasErrors, c.State = c.Validator(c.State)
	if !c.Validation.HasErrors {
		c.State, _ = providers.UpdatePlayerState(c.State)
		c.Stage(c.State)
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
func (c *ShipRepairsController) Validator(state nnsdk.PlayerState) (hasErrors bool, newState nnsdk.PlayerState) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch state.Stage {
	case 0:
		if helpers.InArray(c.Message.Text, []string{
			helpers.Trans("ship.repairs.start"),
		}) {
			state.Stage = 1
			return false, state
		}
	case 1:
		c.Validation.Message = helpers.Trans("wait", state.FinishAt.Format("15:04:05"))
		if state.FinishAt.Before(time.Now()) {
			state.Stage = 2
			return false, state
		}
	}

	return true, state
}

//====================================
// Stage
//====================================
func (c *ShipRepairsController) Stage(state nnsdk.PlayerState) {
	switch state.Stage {
	case 0:
		//====================================
		// Extra data
		//====================================
		needRepair := true
		currentShipRecap := "\n\n"
		eqippedShips, err := providers.GetPlayerShips(helpers.Player, true)
		if err != nil {
			services.ErrorHandler("Cant get equipped player ship", err)
		}

		repairInfo, err := providers.GetShipRepairInfo(eqippedShips[0])
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

		c.Payload.Ship = eqippedShips[0]
		c.Payload.QuantityResources = repairInfo["QuantityResources"].(float64)
		c.Payload.RepairTime = repairInfo["RepairTime"].(float64)
		c.Payload.TypeResources = repairInfo["TypeResources"].(string)

		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)

		c.State, _ = providers.UpdatePlayerState(c.State)

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

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.repairs.info")+currentShipRecap)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		services.SendMessage(msg)

	case 1:

		//====================================
		// Extra data
		//====================================
		// START Repair ship
		responseStart, err := providers.StartShipRepair(c.Payload.Ship)
		if err != nil {
			services.ErrorHandler("Cant repair ship", err)
		}

		recapResourceUsed := fmt.Sprintf("%s\n", helpers.Trans("ship.repairs.used_resources"))
		for resourceID, quantity := range responseStart {
			resource, err := providers.GetResourceByID(resourceID)
			if err != nil {
				services.ErrorHandler("Cant get resource", err)
			}

			recapResourceUsed += fmt.Sprintf("- %s : %v\n", resource.Name, quantity)
		}
		//////////////////////////////////

		// Set timer
		c.State.FinishAt = helpers.GetEndTime(0, int(c.Payload.RepairTime), 0)
		// Stupid poninter stupid json pff
		t := new(bool)
		*t = true
		c.State.ToNotify = t
		c.State.Stage = 1

		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, _ = providers.UpdatePlayerState(c.State)

		msg := services.NewMessage(c.Message.Chat.ID,
			fmt.Sprintf("%s \n %s", recapResourceUsed, helpers.Trans("ship.repairs.reparing", state.FinishAt.Format("15:04:05"))),
		)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)

	case 2:
		// END Repair ship
		_, err := providers.EndShipRepair(c.Payload.Ship)
		if err != nil {
			services.ErrorHandler("Cant repair ship", err)
		}

		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(c.State, helpers.Player)
		//====================================

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.repairs.reparing.finish"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)
	}
}
