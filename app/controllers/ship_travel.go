package controllers

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipTravelController
// ====================================
type ShipTravelController struct {
	BaseController
	Payload struct {
		Ship               *pb.Ship
		StarNearestMapName map[int]string
		StarNearestMapInfo map[int]*pb.GetShipTravelInfo
		StarIDChosen       int
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipTravelController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller:        "route.ship.travel",
		ControllerBlocked: []string{"mission", "hunting"},
		ControllerBack: ControllerBack{
			To:        &ShipController{},
			FromStage: 1,
		},
		Payload: c.Payload,
	}) {
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)

	// Validate
	var hasError bool
	if hasError = c.Validator(); hasError {
		c.Validate()
		return
	}

	// Ok! Run!
	if err = c.Stage(); err != nil {
		panic(err)
	}

	// Completo progressione
	if err = c.Completing(c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *ShipTravelController) Validator() (hasErrors bool) {
	var err error
	switch c.PlayerData.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	// In questo stage non faccio nulla di particolare, verifico solo se ha deciso
	// di avviare una nuova esplorazione
	case 1:
		// A prescindere verifico se il player ha una missione o una caccia attiva
		// tutte le attività di che si svolgono sui pianeti devono essere portati a termine
		for _, state := range c.PlayerData.ActiveStates {
			if helpers.StringInSlice(state.Controller, []string{"route.exploration", "route.hunting"}) {
				c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.travel.error.function_not_completed")
				c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
						),
					),
				)

				return true
			}
		}

		if !helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "ship.travel.start"),
		}) {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
					),
				),
			)

			return true
		}

		return false

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

			return true
		}

		return false

	// In questo stage verificho che l'utente abbia effettivamente aspettato
	// il tempo di attesa necessario al completamento del viaggio
	case 3:
		var finishAt time.Time
		finishAt, err = ptypes.Timestamp(c.PlayerData.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"ship.travel.wait",
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
			return false
		}

		return true
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *ShipTravelController) Stage() (err error) {
	switch c.PlayerData.CurrentState.Stage {

	// Notifico al player la sua posizione e se vuole avviare
	// una nuova esplorazione
	case 0:
		// ****************************
		// Recupero nave attiva de player
		// ****************************
		var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
		rGetPlayerShipEquipped, err = services.NnSDK.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			panic(err)
		}

		// Invio messaggio con recap
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf("%s %s %s",
				helpers.Trans(c.Player.Language.Slug, "ship.travel.info"),
				helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_scanner", rGetPlayerShipEquipped.GetShip().GetShipStats().GetRadar()),
				helpers.Trans(c.Player.Language.Slug, "ship.travel.tip"),
			))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "ship.travel.start")),
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
		c.PlayerData.CurrentState.Stage = 1

	// In questo stage recupero le stelle più vicine disponibili per il player
	case 1:
		// Recupero nave player equipaggiata
		var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
		rGetPlayerShipEquipped, err = services.NnSDK.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		// Recupero informazioni di esplorazione
		var responseTravelInfo *pb.GetShipTravelInfoResponse
		responseTravelInfo, err = services.NnSDK.GetShipTravelInfo(helpers.NewContext(1), &pb.GetShipTravelInfoRequest{
			Ship: rGetPlayerShipEquipped.GetShip(),
		})
		if err != nil {
			return err
		}

		// It's for match with keyboard in validator and needed for next step
		var starNearestMapName = make(map[int]string)
		var starNearestMapInfo = make(map[int]*pb.GetShipTravelInfo)

		var msgNearestStars string
		// Keyboard con riassunto risorse necessarie
		var keyboardRowStars [][]tgbotapi.KeyboardButton
		for _, explorationInfo := range responseTravelInfo.GetInfo() {
			// Se il pianeta è sicuro allora appendo al nome l'icona di riferimento
			planetName := explorationInfo.Planet.Name
			if explorationInfo.Planet.Safe {
				planetName = fmt.Sprintf("%s 🏟", explorationInfo.Planet.Name)
			}

			msgNearestStars += fmt.Sprintf("\n\n🌏 %s\n⏱ %v (%s) ⛽️ -%v%% 🔧 -%v%%",
				planetName,
				explorationInfo.Time/60, helpers.Trans(c.Player.Language.Slug, "hours"),
				explorationInfo.Fuel,
				explorationInfo.Integrity,
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
				helpers.Trans(c.Player.Language.Slug, "ship.travel.research", len(responseTravelInfo.GetInfo())),
				msgNearestStars,
			),
		)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowStars,
		}

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Update state
		c.Payload.Ship = rGetPlayerShipEquipped.GetShip()
		c.Payload.StarNearestMapName = starNearestMapName
		c.Payload.StarNearestMapInfo = starNearestMapInfo
		c.PlayerData.CurrentState.Stage = 2

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
				helpers.Trans(c.Player.Language.Slug, "ship.travel.not_enough_fuel"),
			)

			_, err = services.SendMessage(msg)
			if err != nil {
				return err
			}

			return
		}

		// Setto timer di ritorno
		finishTime := helpers.GetEndTime(0, int(c.Payload.StarNearestMapInfo[chosenStarID].Time), 0)
		c.PlayerData.CurrentState.FinishAt, _ = ptypes.TimestampProto(finishTime)

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.travel.exploring", finishTime.Format("15:04:05 01/02")),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.PlayerData.CurrentState.ToNotify = true
		c.PlayerData.CurrentState.Stage = 3
		c.Payload.StarIDChosen = chosenStarID
		c.ForceBackTo = true

	// Fine esplorazione
	case 3:
		// Costruisco chiamata per aggiornare posizione e scalare il quantitativo
		// di carburante usato
		_, err := services.NnSDK.EndShipTravel(helpers.NewContext(1), &pb.EndShipTravelRequest{
			Integrity: c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Integrity,
			Tank:      c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Fuel,
			PlanetID:  c.Payload.StarNearestMapInfo[c.Payload.StarIDChosen].Planet.ID,
			ShipID:    c.Payload.Ship.ID,
		})
		if err != nil {
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "ship.travel.end"))
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
		c.PlayerData.CurrentState.Completed = true
	}

	return
}
