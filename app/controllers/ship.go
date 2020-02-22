package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipController
// Ogni player ha la possibilità di spostarsi nei diversi pianeti
// del sistema di NoName
// ====================================
type ShipController BaseController

// ====================================
// Handle
// ====================================
func (c *ShipController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	c.Controller = "route.ship"
	c.Player = player
	c.Update = update

	// Recupero nave attiva de player

	var eqippedShips nnsdk.Ships
	eqippedShips, err = providers.GetPlayerShips(c.Player, true)
	if err != nil {
		panic(err)
	}

	var currentShipRecap string
	for _, ship := range eqippedShips {
		currentShipRecap = fmt.Sprintf(
			"\n -%s:\n%s\n\n -%s:\n%s\n\n -%s:\n%s\n\n -%s:\n%v",
			helpers.Trans(c.Player.Language.Slug, "name"), ship.Name,
			helpers.Trans(c.Player.Language.Slug, "category"), ship.ShipCategory.Name,
			helpers.Trans(c.Player.Language.Slug, "rarity"), ship.Rarity.Name,
			helpers.Trans(c.Player.Language.Slug, "integrity"), ship.ShipStats.Integrity,
		)
	}

	// Invio messaggio
	msg := services.NewMessage(c.Update.Message.Chat.ID,
		fmt.Sprintf(
			"%s %s",
			helpers.Trans(c.Player.Language.Slug, "ship.report"),
			currentShipRecap,
		),
	)

	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.exploration")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.repairs")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
		),
	)

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}

// ====================================
// ShipExplorationController
// ====================================
type ShipExplorationController struct {
	BaseController
	Payload struct {
		Ship               nnsdk.Ship
		StarNearestMapName map[int]string
		StarNearestMapInfo map[int]nnsdk.ResponseExplorationInfo
		StarIDChosen       int
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipExplorationController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	c.Controller = "route.ship.exploration"
	c.Player = player
	c.Update = update

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// Validate
	var hasError bool
	hasError, err = c.Validator()
	if err != nil {
		panic(err)
	}

	// Se ritornano degli errori
	if hasError == true {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
		validatorMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
				),
			),
		)

		_, err = services.SendMessage(validatorMsg)
		if err != nil {
			panic(err)
		}

		return
	}

	// Ok! Run!
	err = c.Stage()
	if err != nil {
		panic(err)
	}

	// Aggiorno stato finale
	payloadUpdated, _ := json.Marshal(c.Payload)
	c.State.Payload = string(payloadUpdated)
	c.State, err = providers.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}

	return
}

// ====================================
// Validator
// ====================================
func (c *ShipExplorationController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	switch c.State.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	// In questo stage non faccio nulla di particolare, verifico solo se ha deciso
	// di avviare una nuova esplorazione
	case 1:
		if !helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "ship.exploration.start"),
		}) {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			return true, err
		}

		return false, err

	// In questo stage verifico che il player abbia pasasto la stella vicina
	case 2:
		if !helpers.InArray(c.Update.Message.Text, c.Payload.StarNearestMapName) {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			return true, err
		}

		return false, err

	// In questo stage verificho che l'utente abbia effettivamente aspettato
	// il tempo di attesa necessario al completamento del viaggio
	case 3:
		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"ship.exploration.wait",
			c.State.FinishAt.Format("15:04:05"),
		)

		// Verifico se ha finito il crafting
		if time.Now().After(c.State.FinishAt) {
			return false, err
		}

		return true, err
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *ShipExplorationController) Stage() (err error) {
	switch c.State.Stage {

	// Notifico al player la sua posizione e se vuole avviare
	// una nuova esplorazione
	case 0:
		// Recupero posizione corrente player
		position, err := providers.GetPlayerLastPosition(c.Player)
		if err != nil {
			err = errors.New(fmt.Sprintf("%s %s", "cant get player last position", err))
			return err
		}

		var currentPlayerPositions string
		currentPlayerPositions = fmt.Sprintf(
			"%s \nX: %v \nY: %v \nZ: %v \n",
			helpers.Trans(c.Player.Language.Slug, "ship.exploration.current_position"),
			position.X,
			position.Y,
			position.Z,
		)

		// Invio messaggio con recap
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s\n\n%s",
				helpers.Trans(c.Player.Language.Slug, "ship.exploration.info"),
				currentPlayerPositions,
			),
		)

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "ship.exploration.start")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Avanzo di stato
		c.State.Stage = 1

	// In questo stage recupero le stelle più vicine disponibili per il player
	case 1:
		// Recupero nave player equipaggiata
		eqippedShips, err := providers.GetPlayerShips(c.Player, true)
		if err != nil {
			err = errors.New(fmt.Sprintf("%s %s", "cant get equipped player ship", err))
			return err
		}

		// Recupero informazioni di esplorazione
		explorationInfos, err := providers.GetShipExplorationInfo(eqippedShips[0])
		if err != nil {
			err = errors.New(fmt.Sprintf("%s %s", "cant get player last position", err))
			return err
		}

		// It's for match with keyboard in validator and needed for next step
		var starNearestMapName = make(map[int]string)
		var starNearestMapInfo = make(map[int]nnsdk.ResponseExplorationInfo)

		var msgNearestStars string
		// Keyboard con riassunto risorse necessarie
		var keyboardRowStars [][]tgbotapi.KeyboardButton
		for _, explorationInfo := range explorationInfos {
			msgNearestStars += fmt.Sprintf("\n\n%s:%s\n", helpers.Trans(c.Player.Language.Slug, "name"), explorationInfo.Planet.Name)
			msgNearestStars += fmt.Sprintf("%s:%v\n", helpers.Trans(c.Player.Language.Slug, "ship.exploration.fuel_needed"), explorationInfo.Fuel)
			msgNearestStars += fmt.Sprintf("%s:%v\n", helpers.Trans(c.Player.Language.Slug, "ship.exploration.time_needed"), explorationInfo.Time)
			msgNearestStars += fmt.Sprintf("X: %v \nY: %v \nZ: %v \n\n", explorationInfo.Planet.X, explorationInfo.Planet.Y, explorationInfo.Planet.Z)

			// Aggiungo per la validazione
			starNearestMapName[int(explorationInfo.Planet.ID)] = explorationInfo.Planet.Name
			starNearestMapInfo[int(explorationInfo.Planet.ID)] = explorationInfo

			// Aggiungo stelle alla keyboard
			keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
				explorationInfo.Planet.Name,
			))
			keyboardRowStars = append(keyboardRowStars, keyboardRow)
		}

		keyboardRowStars = append(keyboardRowStars,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s %s",
				helpers.Trans(c.Player.Language.Slug, "ship.exploration.research"),
				msgNearestStars,
			),
		)

		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowStars,
		}

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Update state
		c.Payload.Ship = eqippedShips[0]
		c.Payload.StarNearestMapName = starNearestMapName
		c.Payload.StarNearestMapInfo = starNearestMapInfo
		c.State.Stage = 2

	// Verifico quale stella ha scelto il player e mando messaggio indicando il tempo
	// necessario al suo raggiungimento
	case 2:
		// Filtro e recupero dati stella da raggiungere tramite il messaggio
		var chosenStarID int
		for key, name := range c.Payload.StarNearestMapName {
			if name == c.Update.Message.Text {
				chosenStarID = key
				break
			}
		}

		// Stella non trovata
		if chosenStarID <= 0 {
			err = errors.New("cant get chose star destination")
			return err
		}

		// Setto timer di ritorno
		c.State.FinishAt = helpers.GetEndTime(0, int(c.Payload.StarNearestMapInfo[chosenStarID].Time), 0)

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.exploration.exploring", c.State.FinishAt.Format("15:04:05")),
		)

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.State.ToNotify = helpers.SetTrue()
		c.State.Stage = 3
		c.Payload.StarIDChosen = chosenStarID

	// Fine esplorazione
	case 3:
		// Costruisco chiamata per aggiornare posizione e scalare il quantitativo
		// di carburante usato
		var request nnsdk.RequestExplorationEnd
		request.Position = []float64{
			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Planet.X,
			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Planet.Y,
			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Planet.Z,
		}
		request.Tank = c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Fuel

		_, err := providers.EndShipExploration(c.Payload.Ship, request)
		if err != nil {
			err = errors.New(fmt.Sprintf("%s %s", "cant end exploration", err))
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "ship.exploration.end"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.State.Completed = helpers.SetTrue()
	}

	return
}

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
// 		// Completo lo stato
// 		c.State.Completed = helpers.SetTrue()
// 	}
// }
