package controllers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	MissionTypes = []string{"underground", "surface", "atmosphere"}
)

// ====================================
// MissionController
// ====================================
type MissionController struct {
	BaseController
	Payload struct {
		ExplorationType string // Indica il tipo di esplorazione scelta
		Times           int    // Indica quante volte ha ripetuto
		Dropped         []*pb.DropResourceResponse
		ForcedTime      int // Questo valore serve per forzare le tempistiche
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
		"route.mission",
		c.Payload,
		[]string{"hunting", "ship"},
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

	// In questo stage è necessario controllare che venga scelto
	// un tipo di missione tra quelli disponibili
	case 1:
		// Controllo se il messaggio continene uno dei tipi di missione dichiarati
		for _, missionType := range MissionTypes {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("mission.%s", missionType)) == c.Update.Message.Text {
				return false, err
			}
		}

		return true, err

	// In questo stage andremo a verificare lo stato della missione
	case 2:
		finishAt, err := ptypes.Timestamp(c.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"mission.validator.wait",
			finishAt.Format("15:04:05"),
		)

		// Verifico che l'utente stia accedendo a questa funzionalità solo dopo
		// che abbia finito lo stato attuale e che non abbia raggiunto il limite
		// di volte per il quale è possibile ripetere la stessa azione
		if time.Now().After(finishAt) && c.Payload.Times < 10 {
			c.Payload.Times++

			return false, err
		}

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

	// In questo stage verifico l'azione che vuole intraprendere l'utente
	case 3:
		// Se l'utente decide di continuare/ripetere il ciclo, questo stage si ripete
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "mission.continue") {
			c.CurrentState.FinishAt, _ = ptypes.TimestampProto(helpers.GetEndTime(0, 10*(2*c.Payload.Times), 0))
			c.CurrentState.ToNotify = true

			return false, err

			// Se l'utente invence decide di rientrare e concludere la missione, concludo!
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "mission.comeback") {
			// Passo allo stadio conclusivo
			c.CurrentState.Stage = 4

			return false, err
		}

		return true, err

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
	// Primo avvio di missione, restituisco al player
	// i vari tipi di missioni disponibili
	case 0:
		// Creo messaggio con la lista delle missioni possibili
		var keyboardRows [][]tgbotapi.KeyboardButton
		for _, missionType := range MissionTypes {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("mission.%s", missionType))),
			)

			keyboardRows = append(keyboardRows, keyboardRow)
		}

		// Aggiungo anche abbandona
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
			),
		))

		// Invio messaggi con il tipo di missioni come tastierino
		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "mission.exploration"))
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
		// È il tempo minimo di una missione
		baseMissionTime := 10

		// Verifico se è stato forzato il tempo della prima missione Es. da tutorial
		if c.Payload.ForcedTime > 0 {
			baseMissionTime = c.Payload.ForcedTime
		}

		var endTime time.Time
		endTime = helpers.GetEndTime(0, baseMissionTime, 0)

		// Invio messaggio di attesa
		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"mission.wait",
				endTime.Format("15:04:05"),
			),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return
		}

		// Importo nel payload la scelta di tipologia di missione
		for _, missionType := range MissionTypes {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("mission.%s", missionType)) == c.Update.Message.Text {
				c.Payload.ExplorationType = missionType
				break
			}
		}

		// Avanzo di stato
		c.CurrentState.Stage = 2
		c.CurrentState.ToNotify = true
		c.CurrentState.FinishAt, _ = ptypes.TimestampProto(endTime)
		c.Breaker.ToMenu = true

	// In questo stage recupero quali risorse il player ha recuperato
	// dalla missione e glielo notifico
	case 2:
		// Recupero ultima posizione del player, dando per scontato che sia
		// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
		rGetPlayerLastPosition, err := services.NnSDK.GetPlayerLastPosition(helpers.NewContext(1), &pb.GetPlayerLastPositionRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		// Dalla ultima posizione recupero il pianeta corrente
		rGetPlanetByCoordinate, err := services.NnSDK.GetPlanetByCoordinate(helpers.NewContext(1), &pb.GetPlanetByCoordinateRequest{
			X: rGetPlayerLastPosition.GetPlayerPosition().GetX(),
			Y: rGetPlayerLastPosition.GetPlayerPosition().GetY(),
			Z: rGetPlayerLastPosition.GetPlayerPosition().GetZ(),
		})
		if err != nil {
			return err
		}

		// Recupero drop
		rDropResource, err := services.NnSDK.DropResource(helpers.NewContext(1), &pb.DropResourceRequest{
			TypeExploration: c.Payload.ExplorationType,
			QtyExploration:  int32(c.Payload.Times),
			PlayerID:        c.Player.ID,
			PlanetID:        rGetPlanetByCoordinate.GetPlanet().GetID(),
		})
		if err != nil {
			return err
		}

		// Se ho recuperato il drop lo inserisco nella lista degli elementi droppati
		c.Payload.Dropped = append(c.Payload.Dropped, rDropResource)

		// Invio messaggio di riepilogo con le materie recuperate e chiedo se vuole continuare o ritornare
		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"mission.extraction_recap",
				rDropResource.GetResource().GetName(),
				rDropResource.GetResource().GetRarity().GetName(),
				strings.ToUpper(rDropResource.GetResource().GetRarity().GetSlug()),
				rDropResource.GetQuantity(),
			),
		)
		msg.ParseMode = "markdown"

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "mission.continue")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "mission.comeback")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno lo stato
		c.CurrentState.Stage = 3

	// In questo stage verifico cosa ha scelto di fare il player
	// se ha deciso di continuare allora ritornerò ad uno stato precedente,
	// mentre se ha deciso di concludere andrò avanti di stato
	case 3:
		finishAt, err := ptypes.Timestamp(c.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		// Il player ha scelto di continuare la ricerca
		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"mission.wait",
				finishAt.Format("15:04:05"),
			),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno lo stato
		c.CurrentState.Stage = 2
		c.Breaker.ToMenu = true

	// Ritorno il messaggio con gli elementi droppati
	case 4:
		// Recap delle risorse ricavate da questa missione
		var dropList string
		for _, drop := range c.Payload.Dropped {
			dropList += fmt.Sprintf(
				"- %v x *%s* (%s)\n",
				drop.Quantity,
				drop.Resource.Name,
				strings.ToUpper(drop.Resource.Rarity.Slug),
			)
		}

		// Invio messaggio di chiusura missione
		msg := services.NewMessage(c.Player.ChatID,
			fmt.Sprintf("%s%s",
				helpers.Trans(c.Player.Language.Slug, "mission.extraction_ended"),
				dropList,
			),
		)
		msg.ParseMode = "markdown"

		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiungo le risorse trovare dal player al suo inventario e chiudo
		for _, drop := range c.Payload.Dropped {
			_, err := services.NnSDK.ManagePlayerInventory(helpers.NewContext(1), &pb.ManagePlayerInventoryRequest{
				PlayerID: c.Player.GetID(),
				ItemID:   drop.Resource.ID,
				ItemType: "resources",
				Quantity: drop.Quantity,
			})
			if err != nil {
				return err
			}
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
