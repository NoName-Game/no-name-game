package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipExplorationController
// ====================================
type ShipExplorationController struct {
	BaseController
	Payload struct {
		Ship               *pb.Ship
		StarNearestMapName map[int]string
		StarNearestMapInfo map[int]*pb.GetShipExplorationInfo
		StarIDChosen       int
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipExplorationController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.ship.exploration",
		c.Payload,
		[]string{"mission", "hunting"},
		player,
		update,
	) {
		return
	}

	// Verifico se vuole tornare indietro di stato
	if !proxy {
		if c.BackTo(1, &ShipController{}) {
			return
		}
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

	rUpdatePlayerState, err := services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
		PlayerState: c.State,
	})
	if err != nil {
		panic(err)
	}
	c.State = rUpdatePlayerState.GetPlayerState()

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
		finishAt, err := ptypes.Timestamp(c.State.FinishAt)
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"ship.exploration.wait",
			finishAt.Format("15:04:05 01/02"),
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
		if time.Now().After(finishAt) {
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
		rGetPlayerLastPosition, err := services.NnSDK.GetPlayerLastPosition(helpers.NewContext(1), &pb.GetPlayerLastPositionRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		var currentPlayerPositions string
		currentPlayerPositions = fmt.Sprintf(
			"%s \nX: %v \nY: %v \nZ: %v \n",
			helpers.Trans(c.Player.Language.Slug, "ship.exploration.current_position"),
			rGetPlayerLastPosition.GetPlayerPosition().GetX(),
			rGetPlayerLastPosition.GetPlayerPosition().GetY(),
			rGetPlayerLastPosition.GetPlayerPosition().GetZ(),
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
		rGetPlayerShips, err := services.NnSDK.GetPlayerShips(helpers.NewContext(1), &pb.GetPlayerShipsRequest{
			PlayerID: c.Player.GetID(),
			Equipped: true,
		})
		if err != nil {
			return err
		}

		// Recupero informazioni di esplorazione
		responseExplorationInfo, err := services.NnSDK.GetShipExplorationInfo(helpers.NewContext(1), &pb.GetShipExplorationInfoRequest{
			Ship: rGetPlayerShips.GetShips()[0], // TODO: migliorare
		})
		if err != nil {
			return err
		}

		// It's for match with keyboard in validator and needed for next step
		var starNearestMapName = make(map[int]string)
		var starNearestMapInfo = make(map[int]*pb.GetShipExplorationInfo)

		var msgNearestStars string
		// Keyboard con riassunto risorse necessarie
		var keyboardRowStars [][]tgbotapi.KeyboardButton
		for _, explorationInfo := range responseExplorationInfo.GetShipExplorationInfo {
			// Se il pianeta è sicuro allora appendo al nome l'icona di riferimento
			planetName := explorationInfo.Planet.Name
			if explorationInfo.Planet.Safe {
				planetName = fmt.Sprintf("%s 🏟", explorationInfo.Planet.Name)
			}

			msgNearestStars += fmt.Sprintf("\n\n🌏 %s\n⛽️ -%v%%\nInt️ -%v%%\n⏱ %v (%s)",
				planetName,
				explorationInfo.Fuel,
				explorationInfo.Integrity,
				explorationInfo.Time/60, helpers.Trans(c.Player.Language.Slug, "hours"),
				// explorationInfo.Planet.X, explorationInfo.Planet.Y, explorationInfo.Planet.Z,
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
		c.Payload.Ship = rGetPlayerShips.GetShips()[0]
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
		if c.Payload.StarNearestMapInfo[chosenStarID].Fuel > c.Payload.Ship.ShipStats.Tank {
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
		finishTime := helpers.GetEndTime(0, int(c.Payload.StarNearestMapInfo[chosenStarID].Time), 0)
		c.State.FinishAt, _ = ptypes.TimestampProto(finishTime)

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.exploration.exploring", finishTime.Format("15:04:05 01/02")),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.State.ToNotify = true
		c.State.Stage = 3
		c.Payload.StarIDChosen = chosenStarID
		c.Breaker.ToMenu = true

	// Fine esplorazione
	case 3:
		// Costruisco chiamata per aggiornare posizione e scalare il quantitativo
		// di carburante usato
		// var request nnsdk.ExplorationEndRequest
		// request.Position = []float64{
		// 	c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Planet.X,
		// 	c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Planet.Y,
		// 	c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Planet.Z,
		// }
		// request.Tank = c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Fuel
		// request.Integrity = c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Integrity

		// TODO: da completare

		// _, err := shipProvider.EndShipExploration(c.Payload.Ship, request)
		// if err != nil {
		// 	err = fmt.Errorf("%s %s", "cant end exploration", err)
		// 	return err
		// }

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
		c.State.Completed = true
	}

	return
}
