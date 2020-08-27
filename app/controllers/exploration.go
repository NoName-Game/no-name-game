package controllers

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	ExplorationTypes = []string{"underground", "surface", "atmosphere"}
)

// ====================================
// ExplorationController
// ====================================
type ExplorationController struct {
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
func (c *ExplorationController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller:        "route.exploration",
		ControllerBlocked: []string{"hunting", "ship"},
		ControllerBack: ControllerBack{
			To:        &MenuController{},
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
func (c *ExplorationController) Validator() (hasErrors bool) {
	switch c.PlayerData.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	// In questo stage è necessario controllare che venga scelto
	// un tipo di missione tra quelli disponibili
	case 1:
		// Controllo se il messaggio continene uno dei tipi di missione dichiarati
		for _, missionType := range ExplorationTypes {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("exploration.%s", missionType)) == c.Update.Message.Text {
				return false
			}
		}

		return true

	// In questo stage andremo a verificare lo stato della missione
	case 2:
		var err error
		var finishAt time.Time
		finishAt, err = ptypes.Timestamp(c.PlayerData.CurrentState.GetFinishAt())
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"exploration.validator.wait",
			finishAt.Format("15:04:05"),
		)

		// Verifico che l'utente stia accedendo a questa funzionalità solo dopo
		// che abbia finito lo stato attuale e che non abbia raggiunto il limite
		// di volte per il quale è possibile ripetere la stessa azione
		if time.Now().After(finishAt) && c.Payload.Times < 10 {
			c.Payload.Times++

			return false
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

		return true

	// In questo stage verifico l'azione che vuole intraprendere l'utente
	case 3:
		// Se l'utente decide di continuare/ripetere il ciclo, questo stage si ripete
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "exploration.continue") {
			c.PlayerData.CurrentState.FinishAt, _ = ptypes.TimestampProto(helpers.GetEndTime(0, 10*(2*c.Payload.Times), 0))
			c.PlayerData.CurrentState.ToNotify = true

			return false

			// Se l'utente invence decide di rientrare e concludere la missione, concludo!
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "exploration.comeback") {
			// Passo allo stadio conclusivo
			c.PlayerData.CurrentState.Stage = 4

			return false
		}

		return true

	default:
		// Stato non riconosciuto ritorno errore
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.state")
	}

	// Ritorno errore generico
	return true
}

// ====================================
// Stage
// ====================================
func (c *ExplorationController) Stage() (err error) {
	switch c.PlayerData.CurrentState.Stage {
	// Primo avvio di missione, restituisco al player
	// i vari tipi di missioni disponibili
	case 0:
		// Creo messaggio con la lista delle missioni possibili
		var keyboardRows [][]tgbotapi.KeyboardButton
		for _, missionType := range ExplorationTypes {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("exploration.%s", missionType))),
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
		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "exploration.exploration"))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRows,
			ResizeKeyboard: true,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return
		}

		// Avanzo di stage
		c.PlayerData.CurrentState.Stage = 1

	// In questo stage verrà recuperato il tempo di attesa per il
	// completamnto della missione e notificato al player
	case 1:
		// È il tempo minimo di una missione
		baseExplorationTime := 10

		// Verifico se è stato forzato il tempo della prima missione Es. da tutorial
		if c.Payload.ForcedTime > 0 {
			baseExplorationTime = c.Payload.ForcedTime
		}

		var endTime time.Time
		endTime = helpers.GetEndTime(0, baseExplorationTime, 0)

		// Invio messaggio di attesa
		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"exploration.wait",
				endTime.Format("15:04:05"),
			),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return
		}

		// Importo nel payload la scelta di tipologia di missione
		for _, missionType := range ExplorationTypes {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("exploration.%s", missionType)) == c.Update.Message.Text {
				c.Payload.ExplorationType = missionType
				break
			}
		}

		// Avanzo di stato
		c.PlayerData.CurrentState.Stage = 2
		c.PlayerData.CurrentState.ToNotify = true
		c.PlayerData.CurrentState.FinishAt, _ = ptypes.TimestampProto(endTime)
		c.ForceBackTo = true

	// In questo stage recupero quali risorse il player ha recuperato
	// dalla missione e glielo notifico
	case 2:
		// Recupero ultima posizione del player, dando per scontato che sia
		// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
		var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
		rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		// Recupero drop
		var rDropResource *pb.DropResourceResponse
		rDropResource, err = services.NnSDK.DropResource(helpers.NewContext(1), &pb.DropResourceRequest{
			TypeExploration: c.Payload.ExplorationType,
			QtyExploration:  int32(c.Payload.Times),
			PlayerID:        c.Player.ID,
			PlanetID:        rGetPlayerCurrentPlanet.GetPlanet().GetID(),
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
				"exploration.extraction_recap",
				rDropResource.GetResource().GetName(),
				rDropResource.GetResource().GetRarity().GetName(),
				strings.ToUpper(rDropResource.GetResource().GetRarity().GetSlug()),
				rDropResource.GetQuantity(),
			),
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

		// Aggiorno lo stato
		c.PlayerData.CurrentState.Stage = 3

	// In questo stage verifico cosa ha scelto di fare il player
	// se ha deciso di continuare allora ritornerò ad uno stato precedente,
	// mentre se ha deciso di concludere andrò avanti di stato
	case 3:
		var finishAt time.Time
		finishAt, err = ptypes.Timestamp(c.PlayerData.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		// Il player ha scelto di continuare la ricerca
		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"exploration.wait",
				finishAt.Format("15:04:05"),
			),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno lo stato
		c.PlayerData.CurrentState.Stage = 2
		c.ForceBackTo = true

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
				helpers.Trans(c.Player.Language.Slug, "exploration.extraction_ended"),
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
		c.PlayerData.CurrentState.Completed = true
	}

	return
}
