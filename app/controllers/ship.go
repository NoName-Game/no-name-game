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
			"🚀 %s (%s)\n🏷 %s\n🔧 %v%% (%s)\n⛽ %v%% (%s)",
			ship.Name, ship.Rarity.Slug,
			ship.ShipCategory.Name,
			ship.ShipStats.Integrity, helpers.Trans(c.Player.Language.Slug, "integrity"),
			ship.ShipStats.Tank, helpers.Trans(c.Player.Language.Slug, "fuel"),
		)
	}

	// Invio messaggio
	msg := services.NewMessage(c.Update.Message.Chat.ID,
		fmt.Sprintf(
			"%s:\n\n %s",
			helpers.Trans(c.Player.Language.Slug, "ship.report"),
			currentShipRecap,
		),
	)

	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.exploration")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.rests")),
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
		validatorMsg.ParseMode = "markdown"
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
		// A prescindere verifico se il player ha una missione o una caccia attiva
		// tutte le attività di che si svolgono sui pianeti devono essere portati a termine
		for _, state := range c.Player.States {
			if helpers.StringInSlice(state.Controller, []string{"route.mission", "route.hunting"}) {
				c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.exploration.error.function_not_completed")
				return true, err
			}
		}

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
			c.State.FinishAt.Format("15:04:05 01/02"),
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
			msgNearestStars += fmt.Sprintf("\n\n🌏 %s\n⛽️ -%v%%\n⏱ %v (%s)\nX: %v \nY: %v \nZ: %v",
				explorationInfo.Planet.Name,
				explorationInfo.Fuel,
				explorationInfo.Time/60, helpers.Trans(c.Player.Language.Slug, "hours"),
				explorationInfo.Planet.X, explorationInfo.Planet.Y, explorationInfo.Planet.Z,
			)

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
			helpers.Trans(c.Player.Language.Slug, "ship.exploration.exploring", c.State.FinishAt.Format("15:04:05 01/02")),
		)
		msg.ParseMode = "markdown"

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

// ====================================
// ShipRepairsController
// ====================================
type ShipRepairsController struct {
	BaseController
	Payload struct {
		Ship              nnsdk.Ship
		QuantityResources int
		RepairTime        int
		TypeResources     string
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipRepairsController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	c.Controller = "route.ship.repairs"
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
		validatorMsg.ParseMode = "markdown"
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
func (c *ShipRepairsController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	switch c.State.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.repairs.start") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			return true, err
		}

		return false, err
	case 2:
		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"ship.repairs.wait",
			c.State.FinishAt.Format("15:04:05 01/02"),
		)

		// Verifico se ha finito il crafting
		if time.Now().After(c.State.FinishAt) {
			return false, err
		}
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *ShipRepairsController) Stage() (err error) {
	switch c.State.Stage {

	// In questo riporto al player le risorse e tempistiche necessarie alla riparazione della nave
	case 0:
		// Recupero nave player equipaggiata
		var playerShips nnsdk.Ships
		playerShips, err = providers.GetPlayerShips(c.Player, true)
		if err != nil {
			return err
		}

		// TODO: verificare, dovrebbe recuperarne solo una
		// Recupero name del player
		var playerShip nnsdk.Ship
		playerShip = playerShips[0]

		// Recupero informazioni nave da riparare
		var repairInfo nnsdk.ShipRepairInfoResponse
		repairInfo, err = providers.GetShipRepairInfo(playerShip)
		if err != nil {
			return err
		}

		// Verifico se effettivamente la nave è da riparare
		var shipRecap string
		shipRecap = helpers.Trans(c.Player.Language.Slug, "ship.repairs.info")
		if repairInfo.NeedRepairs {
			shipRecap = fmt.Sprintf("🔧 %v%% (%s)\n%s\n%s ",
				playerShip.ShipStats.Integrity, helpers.Trans(c.Player.Language.Slug, "integrity"),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.time", repairInfo.RepairTime),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.quantity_resources", repairInfo.QuantityResources, repairInfo.TypeResources),
			)
		} else {
			shipRecap = helpers.Trans(c.Player.Language.Slug, "ship.repairs.dont_need")
		}

		// Aggiongo bottone start riparazione
		var keyboardRow [][]tgbotapi.KeyboardButton
		if repairInfo.NeedRepairs {
			newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "ship.repairs.start"),
				),
			)
			keyboardRow = append(keyboardRow, newKeyboardRow)
		}

		// Clear and exit
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
		))

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, shipRecap)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.Payload.Ship = playerShip
		c.Payload.QuantityResources = repairInfo.QuantityResources
		c.Payload.RepairTime = repairInfo.RepairTime
		c.Payload.TypeResources = repairInfo.TypeResources
		c.State.Stage = 1

	// In questo stage avvio effettivamente la riparzione
	case 1:
		// Avvio riparazione nave
		var resourcesUsed []nnsdk.ShipRepairStartResponse
		resourcesUsed, err = providers.StartShipRepair(c.Payload.Ship)
		if err != nil && err.Error() == "not enough resource quantities" {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := services.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.not_enough_resource"),
			)
			_, err = services.SendMessage(errorMsg)
			if err != nil {
				return err
			}

			return
		}

		// Se tutto ok mostro le risorse che vengono consumate per la riparazione
		var recapResourceUsed string
		recapResourceUsed = helpers.Trans(c.Player.Language.Slug, "ship.repairs.used_resources")
		for _, resourceUsed := range resourcesUsed {
			var resource nnsdk.Resource
			resource, err = providers.GetResourceByID(resourceUsed.ResourceID)
			if err != nil {
				return err
			}

			recapResourceUsed += fmt.Sprintf("\n- %s x %v", resource.Name, resourceUsed.Quantity)
		}

		// Setto timer recuperato dalla chiamata delle info
		c.State.FinishAt = helpers.GetEndTime(0, int(c.Payload.RepairTime), 0)

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s \n\n%s",
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.reparing", c.State.FinishAt.Format("15:04:05")),
				recapResourceUsed,
			),
		)

		msg.ParseMode = "markdown"
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
		c.State.Stage = 2
	case 2:
		// Fine riparazione
		err = providers.EndShipRepair(c.Payload.Ship)
		if err != nil {
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.repairs.reparing.finish"),
		)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
				),
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
