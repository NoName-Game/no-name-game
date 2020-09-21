package controllers

import (
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
	Controller
	Payload struct {
		StarNearestMapName map[int]string
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipTravelController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship.travel",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBlocked: []string{"mission", "hunting"},
			ControllerBack: ControllerBack{
				To:        &ShipController{},
				FromStage: 1,
			},
		},
	}) {
		return
	}

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
	if err = c.Completing(&c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *ShipTravelController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	// In questo stage non faccio nulla di particolare, verifico solo se ha deciso
	// di avviare una nuova esplorazione
	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.travel.start") {
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
		var rCheckShipTravel *pb.CheckShipTravelResponse
		if rCheckShipTravel, err = services.NnSDK.CheckShipTravel(helpers.NewContext(1), &pb.CheckShipTravelRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			panic(err)
		}

		// Il crafter sta già portando a terminre un lavoro per questo player
		if !rCheckShipTravel.GetFinishTraveling() {
			var finishAt time.Time
			if finishAt, err = ptypes.Timestamp(rCheckShipTravel.GetTravelingEndTime()); err != nil {
				panic(err)
			}

			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"ship.travel.wait",
				finishAt.Format("15:04:05"),
			)

			return true
		}

		return false
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *ShipTravelController) Stage() (err error) {
	switch c.CurrentState.Stage {

	// Notifico al player la sua posizione e se vuole avviare
	// una nuova esplorazione
	case 0:
		// ****************************
		// Recupero nave attiva de player
		// ****************************
		var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
		if rGetPlayerShipEquipped, err = services.NnSDK.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			panic(err)
		}

		// Invio messaggio con recap
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf("%s %s %s %s %s",
				helpers.Trans(c.Player.Language.Slug, "ship.travel.info"),
				helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_stats"),
				helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_engine", rGetPlayerShipEquipped.GetShip().GetShipStats().GetEngine()),
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

		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Avanzo di stato
		c.CurrentState.Stage = 1

	// In questo stage recupero le stelle più vicine disponibili per il player
	case 1:
		// Recupero informazioni di esplorazione
		var responseTravelInfo *pb.GetShipTravelInfoResponse
		if responseTravelInfo, err = services.NnSDK.GetShipTravelInfo(helpers.NewContext(1), &pb.GetShipTravelInfoRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			return err
		}

		var starNearestMapName = make(map[int]string)
		var starNearestMapInfo = make(map[int]*pb.GetShipTravelInfoResponse_GetShipTravelInfo)
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

		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Update state
		c.Payload.StarNearestMapName = starNearestMapName
		c.CurrentState.Stage = 2

	// Verifico quale stella ha scelto il player e mando messaggio indicando il tempo
	// necessario al suo raggiungimento
	case 2:
		var rStartShipTravel *pb.StartShipTravelResponse
		if rStartShipTravel, err = services.NnSDK.StartShipTravel(helpers.NewContext(1), &pb.StartShipTravelRequest{
			PlayerID:   c.Player.GetID(),
			PlanetName: c.Update.Message.Text,
		}); err != nil {
			return err
		}

		// Recupero orario fine viaggio
		var finishAt time.Time
		if finishAt, err = ptypes.Timestamp(rStartShipTravel.GetTravelingEndTime()); err != nil {
			return
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.travel.exploring", finishAt.Format("15:04:05 01/02")),
		)
		msg.ParseMode = "markdown"
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3
		c.ForceBackTo = true

	// Fine esplorazione
	case 3:
		// Costruisco chiamata per aggiornare posizione e scalare il quantitativo
		// di carburante usato
		if _, err := services.NnSDK.EndShipTravel(helpers.NewContext(1), &pb.EndShipTravelRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "ship.travel.end"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
