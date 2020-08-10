package controllers

import (
	"encoding/json"
	"fmt"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// MissionController
// ====================================
type MissionController struct {
	BaseController
	Payload struct {
		MissionID uint32
	}
}

// ====================================
// Handle
// ====================================
func (c *MissionController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.safeplanet.mission",
		c.Payload,
		[]string{},
		player,
		update,
	) {
		return
	}

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(1, &MenuController{}) {
			return
		}
	}

	// Stato recuperto correttamente
	helpers.UnmarshalPayload(c.CurrentState.Payload, &c.Payload)

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
	c.CurrentState.Payload = string(payloadUpdated)

	rUpdatePlayerState, err := services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
		PlayerState: c.CurrentState,
	})
	if err != nil {
		panic(err)
	}
	c.CurrentState = rUpdatePlayerState.GetPlayerState()

	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *MissionController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		),
	)

	switch c.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	// Verifico se ha scelto di avviare una nuova missione
	case 1:
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.start") == c.Update.Message.Text {
			return false, err
		}

		return true, err

	// In questo stage andremo a verificare lo stato della missione
	case 2:
		var rCheckMission *pb.CheckMissionResponse
		rCheckMission, _ = services.NnSDK.CheckMission(helpers.NewContext(1), &pb.CheckMissionRequest{
			PlayerID:  c.Player.GetID(),
			MissionID: c.Payload.MissionID,
		})

		if rCheckMission.GetCompleted() == false {
			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"safeplanet.mission.check",
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

			return true, err
		}

		return false, err

	default:
		// Stato non riconosciuto ritorno errore
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.state")
	}

	// Ritorno errore generico
	return true, err
}

// ====================================
// Stage
// ====================================
func (c *MissionController) Stage() (err error) {
	switch c.CurrentState.Stage {
	// Primo avvio chiedo al player se vuole avviare una nuova mission
	case 0:
		// Creo messaggio con la lista delle missioni possibili
		var keyboardRows [][]tgbotapi.KeyboardButton
		keyboardRows = append(keyboardRows, []tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.start")),
		})

		// Aggiungo anche abbandona
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
			),
		))

		// Invio messaggi con il tipo di missioni come tastierino
		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.info"))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRows,
			ResizeKeyboard: true,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1

	// In questo stage verrà recuperato il tempo di attesa per il
	// completamnto della missione e notificato al player
	case 1:
		// Chiamo il ws e recupero il tipo di missione da effettuare
		// attraverso il tipo di missione costruisco il corpo del messaggio
		var rGetMission *pb.GetMissionResponse
		rGetMission, err = services.NnSDK.GetMission(helpers.NewContext(1), &pb.GetMissionRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			panic(err)
		}

		// In base alla categoria della missione costruisco il messaggio
		var missionRecap string
		switch rGetMission.GetMission().GetMissionCategory().GetSlug() {
		case "resources_finding":
			var missionPayload *pb.MissionResourcesFinding
			helpers.UnmarshalPayload(rGetMission.GetMission().GetPayload(), &missionPayload)

			// Recupero enitità risorsa da cercare
			var rGetResourceByID *pb.GetResourceByIDResponse
			rGetResourceByID, err = services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: missionPayload.GetResourceID(),
			})
			if err != nil {
				panic(err)
			}

			missionRecap += fmt.Sprintf("Sei stato affidato ad una missione di tipologia: *%v*\nDevi recuperare *%v* quantità di 💠 *%v*. Buona Fortuna.",
				rGetMission.GetMission().GetMissionCategory().GetName(),
				missionPayload.GetResourceQty(),
				rGetResourceByID.GetResource().GetName(),
			)
		}

		// Invio messaggio di attesa
		msg := services.NewMessage(c.Player.ChatID,
			missionRecap,
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return
		}

		// Avanzo di stato
		c.Payload.MissionID = rGetMission.GetMission().GetID()
		c.CurrentState.Stage = 2
		c.Breaker.ToMenu = true

	case 2:
		// Effettuo chiamata al WS per verificare che abbia completato la missione
		// e recupero drop
		msg := services.NewMessage(c.Player.ChatID,
			"Bravo hai completato tutto, grazie ciao",
		)
		msg.ParseMode = "markdown"

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "exploration.continue")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "exploration.comeback")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
