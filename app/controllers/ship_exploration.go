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
// ShipExplorationController
// ====================================
type ShipExplorationController struct {
	BaseController
	Payload struct {
		Ship               nnsdk.Ship
		StarNearestMapName map[int]string
		StarNearestMapInfo map[int]nnsdk.ExplorationInfoResponse
		StarIDChosen       int
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipExplorationController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	var playerStateProvider providers.PlayerStateProvider

	c.Controller = "route.ship.exploration"
	c.Player = player
	c.Update = update

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	if c.Clear() {
		return
	}

	// Verifico se vuole tornare indietro di stato
	if c.BackTo(1) {
		new(ShipController).Handle(c.Player, c.Update)
		return
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
	if hasError {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
		validatorMsg.ParseMode = "markdown"
		validatorMsg.ReplyMarkup = c.Validation.ReplyKeyboard

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
	c.State, err = playerStateProvider.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *ShipExplorationController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		),
	)

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
				c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
						),
					),
				)

				return true, err
			}
		}

		if !helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "ship.exploration.start"),
		}) {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
					),
				),
			)

			return true, err
		}

		return false, err

	// In questo stage verifico che il player abbia pasasto la stella vicina
	case 2:
		if !helpers.InArray(c.Update.Message.Text, c.Payload.StarNearestMapName) {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
					),
				),
			)

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

		// Aggiungo anche abbandona
		c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.continue"),
				),
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
				),
			),
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
	var playerProvider providers.PlayerProvider
	var shipProvider providers.ShipProvider

	switch c.State.Stage {

	// Notifico al player la sua posizione e se vuole avviare
	// una nuova esplorazione
	case 0:
		// Recupero posizione corrente player
		var position nnsdk.PlayerPosition
		position, err = playerProvider.GetPlayerLastPosition(c.Player)
		if err != nil {
			err = fmt.Errorf("%s %s", "cant get player last position", err)
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
		var eqippedShips nnsdk.Ships
		eqippedShips, err = playerProvider.GetPlayerShips(c.Player, true)
		if err != nil {
			err = fmt.Errorf("%s %s", "cant get equipped player ship", err)
			return err
		}

		// Recupero informazioni di esplorazione
		var explorationInfos []nnsdk.ExplorationInfoResponse
		explorationInfos, err = shipProvider.GetShipExplorationInfo(eqippedShips[0])
		if err != nil {
			err = fmt.Errorf("%s %s", "cant get player last position", err)
			return err
		}

		// It's for match with keyboard in validator and needed for next step
		var starNearestMapName = make(map[int]string)
		var starNearestMapInfo = make(map[int]nnsdk.ExplorationInfoResponse)

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
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
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

		// Verifico se la nave del player ha abbastanza carburante per raggiungere la stella
		if c.Payload.StarNearestMapInfo[chosenStarID].Fuel > *c.Payload.Ship.ShipStats.Tank {
			msg := services.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "ship.exploration.not_enough_fuel"),
			)

			_, err = services.SendMessage(msg)
			if err != nil {
				return err
			}

			return
		}

		// Setto timer di ritorno
		c.State.FinishAt = helpers.GetEndTime(0, int(c.Payload.StarNearestMapInfo[chosenStarID].Time), 0)

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.exploration.exploring", c.State.FinishAt.Format("15:04:05 01/02")),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		*c.State.ToNotify = true
		c.State.Stage = 3
		c.Payload.StarIDChosen = chosenStarID
		c.Breaker.ToMenu = true

	// Fine esplorazione
	case 3:
		// Costruisco chiamata per aggiornare posizione e scalare il quantitativo
		// di carburante usato
		var request nnsdk.ExplorationEndRequest
		request.Position = []float64{
			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Planet.X,
			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Planet.Y,
			c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Planet.Z,
		}
		request.Tank = c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Fuel

		_, err := shipProvider.EndShipExploration(c.Payload.Ship, request)
		if err != nil {
			err = fmt.Errorf("%s %s", "cant end exploration", err)
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "ship.exploration.end"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		*c.State.Completed = true
	}

	return
}
