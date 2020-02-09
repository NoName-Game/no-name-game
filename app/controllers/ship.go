package controllers

//
// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"strconv"
// 	"time"
//
// 	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
//
// 	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
// 	"bitbucket.org/no-name-game/nn-telegram/app/providers"
// 	"bitbucket.org/no-name-game/nn-telegram/services"
// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
// )
//
// //====================================
// // ShipController
// //====================================
// type ShipController BaseController
//
// //====================================
// // Handle
// //====================================
// func (c *ShipController) Handle(update tgbotapi.Update) {
// 	c.Message = update.Message
//
// 	// Recuper nave attiva de player
// 	currentShipRecap := "\n\n"
// 	eqippedShips, err := providers.GetPlayerShips(helpers.Player, true)
// 	if err != nil {
// 		services.ErrorHandler("Cant get equipped player ship", err)
// 	}
//
// 	for _, ship := range eqippedShips {
// 		currentShipRecap += helpers.Trans("name") + ": " + ship.Name + "\n"
// 		currentShipRecap += helpers.Trans("category") + ": " + ship.ShipCategory.Name + "\n"
// 		currentShipRecap += helpers.Trans("rarity") + ": " + ship.Rarity.Name + "\n"
// 		currentShipRecap += helpers.Trans("integrity") + ": " + strconv.FormatUint(uint64(ship.ShipStats.Integrity), 10) + "\n"
// 	}
//
// 	// Invio messaggio
// 	msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.report")+currentShipRecap)
// 	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 		tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton(helpers.Trans("route.ship.exploration")),
// 		),
// 		tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton(helpers.Trans("route.ship.repairs")),
// 		),
// 		tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
// 		),
// 	)
//
// 	services.SendMessage(msg)
// }
//
// //====================================
// // ShipExplorationController
// //====================================
// type ShipExplorationController struct {
// 	BaseController
// 	Payload struct {
// 		Ship               nnsdk.Ship
// 		StarNearestMapName map[int]string
// 		StarNearestMapInfo map[int]providers.ResponseExplorationInfo
// 		StarIDChosen       int
// 	}
// }
//
// //====================================
// // Handle
// //====================================
// func (c *ShipExplorationController) Handle(update tgbotapi.Update) {
// 	// Current Controller instance
// 	var err error
// 	var isNewState bool
// 	c.RouteName, c.Update, c.Message = "route.ship.exploration", update, update.Message
//
// 	// Check current state for this routes
// 	c.State, isNewState = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)
//
// 	// Set and load payload
// 	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)
//
// 	// It's first message
// 	if isNewState {
// 		c.Stage()
// 		return
// 	}
//
// 	// Go to validator
// 	if !c.Validator() {
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
//
// 		// Ok! Run!
// 		c.Stage()
// 		return
// 	}
//
// 	// Validator goes errors
// 	validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
// 	services.SendMessage(validatorMsg)
// 	return
// }
//
// //====================================
// // Validator
// //====================================
// func (c *ShipExplorationController) Validator() (hasErrors bool) {
// 	c.Validation.Message = helpers.Trans("validationMessage")
//
// 	switch c.State.Stage {
// 	case 0:
// 		if helpers.InArray(c.Message.Text, []string{
// 			helpers.Trans("ship.exploration.start"),
// 		}) {
// 			c.State.Stage = 1
// 			return false
// 		}
// 	case 1:
// 		if helpers.InArray(c.Message.Text, c.Payload.StarNearestMapName) {
// 			c.State.Stage = 2
// 			return false
// 		}
// 	case 2:
// 		c.Validation.Message = helpers.Trans("wait", c.State.FinishAt.Format("15:04:05"))
// 		if time.Now().After(c.State.FinishAt) {
// 			c.State.Stage = 3
// 			return false
// 		}
// 	}
//
// 	return true
// }
//
// //====================================
// // Stage
// //====================================
// func (c *ShipExplorationController) Stage() {
// 	var err error
//
// 	switch c.State.Stage {
// 	case 0:
// 		// TODO: Verificare se esiste in helpers.Player
// 		// Recupero posizione corrente player
// 		currentPlayerPositions := "\n\n"
// 		position, err := providers.GetPlayerLastPosition(helpers.Player)
// 		if err != nil {
// 			services.ErrorHandler("Cant get player last position", err)
// 		}
//
// 		currentPlayerPositions += fmt.Sprintf("%s \nX: %v \nY: %v \nZ: %v \n", helpers.Trans("ship.exploration.current_position"), position.X, position.Y, position.Z)
//
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.exploration.info")+currentPlayerPositions)
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("ship.exploration.start")),
// 			),
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
// 			),
// 		)
// 		services.SendMessage(msg)
// 	case 1:
// 		// TODO: Verificare se esiste in helpers.Player
// 		// Recupero nave player equipaggiata
// 		eqippedShips, err := providers.GetPlayerShips(helpers.Player, true)
// 		if err != nil {
// 			services.ErrorHandler("Cant get equipped player ship", err)
// 		}
//
// 		// Recupero informazioni di esplorazione
// 		msgNearestStars := "\n\n"
// 		explorationInfos, err := providers.GetShipExplorationInfo(eqippedShips[0])
// 		if err != nil {
// 			services.ErrorHandler("Cant get player last position", err)
// 		}
//
// 		// It's for match with keyboard in validator and needed for next step
// 		var starNearestMapName = make(map[int]string)
// 		var starNearestMapInfo = make(map[int]providers.ResponseExplorationInfo)
//
// 		// Keyboard con e riassunto risorse necessarie
// 		var keyboardRowStars [][]tgbotapi.KeyboardButton
// 		for _, explorationInfo := range explorationInfos {
// 			msgNearestStars += fmt.Sprintf("%s:%s\n", helpers.Trans("name"), explorationInfo.Star.Name)
// 			msgNearestStars += fmt.Sprintf("%s:%v\n", helpers.Trans("ship.exploration.fuel_needed"), explorationInfo.Fuel)
// 			msgNearestStars += fmt.Sprintf("%s:%v\n", helpers.Trans("ship.exploration.time_needed"), explorationInfo.Time)
// 			msgNearestStars += fmt.Sprintf("X: %v \nY: %v \nZ: %v \n\n", explorationInfo.Star.X, explorationInfo.Star.Y, explorationInfo.Star.Z)
//
// 			// Add for validation and next step
// 			starNearestMapName[int(explorationInfo.Star.ID)] = explorationInfo.Star.Name
// 			starNearestMapInfo[int(explorationInfo.Star.ID)] = explorationInfo
//
// 			// Add for keyboard
// 			keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
// 				explorationInfo.Star.Name,
// 			))
// 			keyboardRowStars = append(keyboardRowStars, keyboardRow)
// 		}
//
// 		// Clear and exit
// 		keyboardRowStars = append(keyboardRowStars,
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
// 			),
// 		)
//
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.exploration.research")+msgNearestStars)
// 		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
// 			ResizeKeyboard: true,
// 			Keyboard:       keyboardRowStars,
// 		}
// 		services.SendMessage(msg)
//
// 		// Update state
// 		c.Payload.Ship = eqippedShips[0]
// 		c.Payload.StarNearestMapName = starNearestMapName
// 		c.Payload.StarNearestMapInfo = starNearestMapInfo
// 		payloadUpdated, _ := json.Marshal(c.Payload)
// 		c.State.Payload = string(payloadUpdated)
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
// 	case 2:
// 		// Filtro e recupero dati stella da raggiungere tramite il messaggio
// 		var starIDchosen int
// 		for key, name := range c.Payload.StarNearestMapName {
// 			if name == c.Message.Text {
// 				starIDchosen = key
// 				break
// 			}
// 		}
//
// 		// Stella non trovata
// 		if starIDchosen <= 0 {
// 			services.ErrorHandler("Cant get chose star destination", errors.New("Cant get chose star destination"))
// 		}
//
// 		// Setto timer di ritorno
// 		c.State.FinishAt = helpers.GetEndTime(0, int(c.Payload.StarNearestMapInfo[starIDchosen].Time), 0)
//
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID,
// 			fmt.Sprintf("%s \n", helpers.Trans("ship.exploration.exploring", c.State.FinishAt.Format("15:04:05"))),
// 		)
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 			),
// 		)
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		c.State.ToNotify = helpers.SetTrue()
// 		c.State.Stage = 2
// 		c.Payload.StarIDChosen = starIDchosen
//
// 		payloadUpdated, _ := json.Marshal(c.Payload)
// 		c.State.Payload = string(payloadUpdated)
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
// 	case 3:
// 		// End exploration
// 		var request providers.RequestExplorationEnd
// 		request.Position = []float64{
// 			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Star.X,
// 			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Star.Y,
// 			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Star.Z,
// 		}
// 		request.Tank = c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Fuel
//
// 		_, err := providers.EndShipExploration(c.Payload.Ship, request)
// 		if err != nil {
// 			services.ErrorHandler("Cant end exploration", err)
// 		}
//
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.exploration.end"))
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
// 			),
// 		)
// 		services.SendMessage(msg)
//
// 		//====================================
// 		// COMPLETE!
// 		//====================================
// 		helpers.FinishAndCompleteState(c.State, helpers.Player)
// 		//====================================
// 	}
// }
//
// //====================================
// // ShipRepairsController
// //====================================
// type ShipRepairsController struct {
// 	BaseController
// 	Payload struct {
// 		Ship              nnsdk.Ship
// 		QuantityResources float64
// 		RepairTime        float64
// 		TypeResources     string
// 	}
// }
//
// //====================================
// // Handle
// //====================================
// func (c *ShipRepairsController) Handle(update tgbotapi.Update) {
// 	// Current Controller instance
// 	var err error
// 	var isNewState bool
// 	c.RouteName, c.Update, c.Message = "route.ship.repairs", update, update.Message
//
// 	// Check current state for this routes
// 	c.State, isNewState = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)
//
// 	// Set and load payload
// 	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)
//
// 	// It's first message
// 	if isNewState {
// 		c.Stage()
// 		return
// 	}
//
// 	// Go to validator
// 	if !c.Validator() {
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
//
// 		// Ok! Run!
// 		c.Stage()
// 		return
// 	}
//
// 	// Validator goes errors
// 	validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
// 	services.SendMessage(validatorMsg)
// 	return
// }
//
// //====================================
// // Validator
// //====================================
// func (c *ShipRepairsController) Validator() (hasErrors bool) {
// 	c.Validation.Message = helpers.Trans("validationMessage")
//
// 	switch c.State.Stage {
// 	case 0:
// 		if helpers.InArray(c.Message.Text, []string{
// 			helpers.Trans("ship.repairs.start"),
// 		}) {
// 			c.State.Stage = 1
// 			return false
// 		}
// 	case 1:
// 		c.Validation.Message = helpers.Trans("wait", c.State.FinishAt.Format("15:04:05"))
// 		if c.State.FinishAt.Before(time.Now()) {
// 			c.State.Stage = 2
// 			return false
// 		}
// 	}
//
// 	return true
// }
//
// //====================================
// // Stage
// //====================================
// func (c *ShipRepairsController) Stage() {
// 	switch c.State.Stage {
// 	case 0:
// 		needRepair := true
//
// 		// Recupero nave player equipaggiata
// 		currentShipRecap := "\n\n"
// 		eqippedShips, err := providers.GetPlayerShips(helpers.Player, true)
// 		if err != nil {
// 			services.ErrorHandler("Cant get equipped player ship", err)
// 		}
//
// 		// Recupero informazioni nave da riparare
// 		repairInfo, err := providers.GetShipRepairInfo(eqippedShips[0])
// 		if err != nil {
// 			services.ErrorHandler("Cant get ship repair info", err)
// 		}
//
// 		// Verifico se effettivamente la nave è da riparare
// 		if repairInfo["QuantityResources"].(float64) <= 0 {
// 			needRepair = false
// 			currentShipRecap += helpers.Trans("ship.repairs.dont_need")
// 		} else {
// 			currentShipRecap += fmt.Sprintf("%s: %v\n", helpers.Trans("integrity"), eqippedShips[0].ShipStats.Integrity)
// 			currentShipRecap += fmt.Sprintf("%s: %v %s\n", helpers.Trans("ship.repairs.time"), repairInfo["RepairTime"], helpers.Trans("minutes"))
// 			currentShipRecap += fmt.Sprintf("%s: %v (%v)\n", helpers.Trans("ship.repairs.quantity_resources"), repairInfo["QuantityResources"], repairInfo["TypeResources"])
// 		}
//
// 		// Aggiongo bottone start riparazione
// 		var keyboardRow [][]tgbotapi.KeyboardButton
// 		if needRepair {
// 			newKeyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("ship.repairs.start")))
// 			keyboardRow = append(keyboardRow, newKeyboardRow)
// 		}
//
// 		// Clear and exit
// 		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
// 		))
//
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.repairs.info")+currentShipRecap)
// 		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
// 			ResizeKeyboard: true,
// 			Keyboard:       keyboardRow,
// 		}
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		c.Payload.Ship = eqippedShips[0]
// 		c.Payload.QuantityResources = repairInfo["QuantityResources"].(float64)
// 		c.Payload.RepairTime = repairInfo["RepairTime"].(float64)
// 		c.Payload.TypeResources = repairInfo["TypeResources"].(string)
//
// 		payloadUpdated, _ := json.Marshal(c.Payload)
// 		c.State.Payload = string(payloadUpdated)
//
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
// 	case 1:
// 		// Avvio riparazione nave
// 		responseStart, err := providers.StartShipRepair(c.Payload.Ship)
// 		if err != nil {
// 			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
// 			errorMsg := services.NewMessage(c.Message.Chat.ID,
// 				fmt.Sprintf("%s", err),
// 			)
// 			services.SendMessage(errorMsg)
// 			return
// 		}
//
// 		recapResourceUsed := fmt.Sprintf("%s\n", helpers.Trans("ship.repairs.used_resources"))
// 		for resourceID, quantity := range responseStart {
// 			resource, err := providers.GetResourceByID(resourceID)
// 			if err != nil {
// 				services.ErrorHandler("Cant get resource", err)
// 			}
//
// 			recapResourceUsed += fmt.Sprintf("- %s : %v\n", resource.Item.Name, quantity)
// 		}
//
// 		// Setto timer
// 		c.State.FinishAt = helpers.GetEndTime(0, int(c.Payload.RepairTime), 0)
//
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID,
// 			fmt.Sprintf("%s \n %s", recapResourceUsed, helpers.Trans("ship.repairs.reparing", c.State.FinishAt.Format("15:04:05"))),
// 		)
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 			),
// 		)
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		c.State.ToNotify = helpers.SetTrue()
// 		c.State.Stage = 1
//
// 		payloadUpdated, _ := json.Marshal(c.Payload)
// 		c.State.Payload = string(payloadUpdated)
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
// 	case 2:
// 		// Fine riparazione
// 		_, err := providers.EndShipRepair(c.Payload.Ship)
// 		if err != nil {
// 			services.ErrorHandler("Cant repair ship", err)
// 		}
//
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("ship.repairs.reparing.finish"))
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
// 			),
// 		)
// 		services.SendMessage(msg)
//
// 		//====================================
// 		// COMPLETE!
// 		//====================================
// 		helpers.FinishAndCompleteState(c.State, helpers.Player)
// 		//====================================
// 	}
// }
