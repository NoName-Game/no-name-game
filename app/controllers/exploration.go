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

	log "github.com/sirupsen/logrus"
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
	// helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)

	// Validate
	var hasError bool
	if hasError = c.Validator(); hasError {
		c.Validate()
		return
	}

	log.Info("Prima di stage: ", c.CurrentState.Stage)

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
func (c *ExplorationController) ValidatorNEW() (hasErrors bool) {
	var err error

	var rExplorationCheck *pb.ExplorationCheckResponse
	if rExplorationCheck, err = services.NnSDK.ExplorationCheck(helpers.NewContext(1), &pb.ExplorationCheckRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		panic(err)
	}

	// Se il player NON si trova in esplorazione
	if !rExplorationCheck.GetInExploration() {
		// Se NON è in missione

		// Verifico se viene specifica una missione in particolare

		// Recupero tutte le categorie di esplorazione possibili
		var rGetAllExplorationCategories *pb.GetAllExplorationCategoriesResponse
		if rGetAllExplorationCategories, err = services.NnSDK.GetAllExplorationCategories(helpers.NewContext(1), &pb.GetAllExplorationCategoriesRequest{}); err != nil {
			panic(err)
		}

		// Controllo se il messaggio continene uno dei tipi di missione dichiarati
		for _, missionType := range rGetAllExplorationCategories.GetExplorationCategories() {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("exploration.%s", missionType.GetSlug())) == c.Update.Message.Text {
				return false
			}
		}
	}

	// Il Player deve terminare prima l'esplorazione in corso
	if !rExplorationCheck.GetFinishExploration() {
		var finishAt time.Time
		finishAt, err = ptypes.Timestamp(rExplorationCheck.GetExplorationEndTime())
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"exploration.validator.wait",
			finishAt.Format("15:04:05"),
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

		return true
	}

	return false
}

// ====================================
// Validator
// ====================================
func (c *ExplorationController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	// In questo stage è necessario controllare che venga scelto
	// un tipo di missione tra quelli disponibili
	case 1:
		// Recupero tutte le categorie di esplorazione possibili
		var rGetAllExplorationCategories *pb.GetAllExplorationCategoriesResponse
		if rGetAllExplorationCategories, err = services.NnSDK.GetAllExplorationCategories(helpers.NewContext(1), &pb.GetAllExplorationCategoriesRequest{}); err != nil {
			panic(err)
		}

		// Controllo se il messaggio continene uno dei tipi di missione dichiarati
		for _, missionType := range rGetAllExplorationCategories.GetExplorationCategories() {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("exploration.%s", missionType.GetSlug())) == c.Update.Message.Text {
				return false
			}
		}

		return true

	// In questo stage andremo a verificare lo stato della missione
	case 2:
		var rExplorationCheck *pb.ExplorationCheckResponse
		if rExplorationCheck, err = services.NnSDK.ExplorationCheck(helpers.NewContext(1), &pb.ExplorationCheckRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			panic(err)
		}

		// Il Player deve terminare prima l'esplorazione in corso
		if !rExplorationCheck.GetFinishExploration() {
			var finishAt time.Time
			finishAt, err = ptypes.Timestamp(rExplorationCheck.GetExplorationEndTime())
			if err != nil {
				panic(err)
			}

			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"exploration.validator.wait",
				finishAt.Format("15:04:05"),
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

			return true
		}

		return false
	// In questo stage verifico l'azione che vuole intraprendere l'utente
	case 3:
		// Se l'utente decide di continuare/ripetere il ciclo, questo stage si ripete
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "exploration.continue") {
			// c.PlayerData.CurrentState.FinishAt, _ = ptypes.TimestampProto(helpers.GetEndTime(0, 10*(2*c.Payload.Times), 0))
			// c.PlayerData.CurrentState.ToNotify = true

			return false

			// Se l'utente invence decide di rientrare e concludere la missione, concludo!
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "exploration.comeback") {
			// Passo allo stadio conclusivo
			c.CurrentState.Stage = 4

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
	switch c.CurrentState.Stage {
	// Primo avvio di missione, restituisco al player
	// i vari tipi di missioni disponibili
	case 0:
		// Recupero tutte le categorie di esplorazione possibili
		var rGetAllExplorationCategories *pb.GetAllExplorationCategoriesResponse
		rGetAllExplorationCategories, err = services.NnSDK.GetAllExplorationCategories(helpers.NewContext(1), &pb.GetAllExplorationCategoriesRequest{})
		if err != nil {
			return err
		}

		// Creo messaggio con la lista delle missioni possibili
		var keyboardRows [][]tgbotapi.KeyboardButton
		for _, missionType := range rGetAllExplorationCategories.GetExplorationCategories() {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("exploration.%s", missionType.GetSlug()))),
			)

			keyboardRows = append(keyboardRows, keyboardRow)
		}

		// Aggiungo anche abbandona
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
			),
		))

		// Costruisco risposta
		var messageExploration string
		messageExploration = helpers.Trans(c.Player.Language.Slug, "exploration.exploration")

		// Recupero posizione player
		var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
		rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
			PlayerID: c.Player.ID,
		})
		if err != nil {
			return err
		}

		// Verifico se sono conquistatore
		var rGetCurrentConquerorByPlanetID *pb.GetCurrentConquerorByPlanetIDResponse
		rGetCurrentConquerorByPlanetID, err = services.NnSDK.GetCurrentConquerorByPlanetID(helpers.NewContext(1), &pb.GetCurrentConquerorByPlanetIDRequest{
			PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
		})
		if err != nil {
			return err
		}

		// Verifico se il player è un conquistatore
		if c.Player.ID == rGetCurrentConquerorByPlanetID.GetPlayer().GetID() {
			messageExploration += helpers.Trans(c.Player.Language.Slug, "exploration.conqueror_bonus")
		}

		// Invio messaggi con il tipo di missioni come tastierino
		msg := services.NewMessage(c.Player.ChatID, messageExploration)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRows,
			ResizeKeyboard: true,
		}
		if _, err = services.SendMessage(msg); err != nil {
			return
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1

	// In questo stage verrà recuperato il tempo di attesa per il
	// completamnto della missione e notificato al player
	case 1:
		// Recupero tutte le categorie di esplorazione possibili
		var rGetAllExplorationCategories *pb.GetAllExplorationCategoriesResponse
		rGetAllExplorationCategories, err = services.NnSDK.GetAllExplorationCategories(helpers.NewContext(1), &pb.GetAllExplorationCategoriesRequest{})
		if err != nil {
			return err
		}

		// Recupero dal messaggio quale esplorazione vuole effettuare il player
		var explorationChoiced string
		for _, missionType := range rGetAllExplorationCategories.GetExplorationCategories() {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("exploration.%s", missionType.GetSlug())) == c.Update.Message.Text {
				explorationChoiced = missionType.GetSlug()
				break
			}
		}

		// Avvio nuova esplorazione
		var rExplorationStart *pb.ExplorationStartResponse
		if rExplorationStart, err = services.NnSDK.ExplorationStart(helpers.NewContext(1), &pb.ExplorationStartRequest{
			PlayerID:                c.Player.ID,
			ExplorationCategorySlug: explorationChoiced,
		}); err != nil {
			return err
		}

		// Converto finishAt in formato Time
		var finishAt time.Time
		if finishAt, err = ptypes.Timestamp(rExplorationStart.GetFinishAt()); err != nil {
			return err
		}

		// Invio messaggio di attesa
		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"exploration.wait",
				finishAt.Format("15:04:05"),
			),
		)
		msg.ParseMode = "markdown"
		if _, err = services.SendMessage(msg); err != nil {
			return
		}

		// Avanzo di stage
		c.CurrentState.Stage = 2
		c.ForceBackTo = true

	// In questo stage recupero quali risorse il player ha recuperato
	// dalla missione e glielo notifico
	case 2:
		// Recupero ultima posizione del player, dando per scontato che sia
		// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
		// var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
		// rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
		// 	PlayerID: c.Player.GetID(),
		// })
		// if err != nil {
		// 	return err
		// }
		//
		// // Recupero drop
		// var rDropResource *pb.DropResourceResponse
		// rDropResource, err = services.NnSDK.DropResource(helpers.NewContext(1), &pb.DropResourceRequest{
		// 	TypeExploration: c.Payload.ExplorationType,
		// 	QtyExploration:  int32(c.Payload.Times),
		// 	PlayerID:        c.Player.ID,
		// 	PlanetID:        rGetPlayerCurrentPlanet.GetPlanet().GetID(),
		// })
		// if err != nil {
		// 	return err
		// }
		//
		// // Se ho recuperato il drop lo inserisco nella lista degli elementi droppati
		// c.Payload.Dropped = append(c.Payload.Dropped, rDropResource)

		var rExplorationCheck *pb.ExplorationCheckResponse
		if rExplorationCheck, err = services.NnSDK.ExplorationCheck(helpers.NewContext(1), &pb.ExplorationCheckRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			panic(err)
		}

		// Invio messaggio di riepilogo con le materie recuperate e chiedo se vuole continuare o ritornare
		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"exploration.extraction_recap",
				rExplorationCheck.GetDropResult().GetResource().GetName(),
				rExplorationCheck.GetDropResult().GetResource().GetRarity().GetName(),
				strings.ToUpper(rExplorationCheck.GetDropResult().GetResource().GetRarity().GetSlug()),
				rExplorationCheck.GetDropResult().GetQuantity(),
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
		// c.PlayerData.CurrentState.Stage = 3
		c.CurrentState.Stage = 3

	// In questo stage verifico cosa ha scelto di fare il player
	// se ha deciso di continuare allora ritornerò ad uno stato precedente,
	// mentre se ha deciso di concludere andrò avanti di stato
	case 3:
		// var finishAt time.Time
		// finishAt, err = ptypes.Timestamp(c.PlayerData.CurrentState.FinishAt)
		// if err != nil {
		// 	panic(err)
		// }

		// Continua esplorazione
		var rExplorationContinue *pb.ExplorationContinueResponse
		if rExplorationContinue, err = services.NnSDK.ExplorationContinue(helpers.NewContext(1), &pb.ExplorationContinueRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			panic(err)
		}

		// Converto finishAt in formato Time
		var finishAt time.Time
		if finishAt, err = ptypes.Timestamp(rExplorationContinue.GetFinishAt()); err != nil {
			return err
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
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorno lo stato
		// c.PlayerData.CurrentState.Stage = 2

		c.CurrentState.Stage = 2
		c.ForceBackTo = true

	// Ritorno il messaggio con gli elementi droppati
	case 4:
		var rExplorationEnd *pb.ExplorationEndResponse
		if rExplorationEnd, err = services.NnSDK.ExplorationEnd(helpers.NewContext(1), &pb.ExplorationEndRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			panic(err)
		}

		// Recap delle risorse ricavate da questa missione
		var dropList string
		for _, drop := range rExplorationEnd.GetDropResults() {
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
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiungo le risorse trovare dal player al suo inventario e chiudo
		// for _, drop := range c.Payload.Dropped {
		// 	_, err := services.NnSDK.ManagePlayerInventory(helpers.NewContext(1), &pb.ManagePlayerInventoryRequest{
		// 		PlayerID: c.Player.GetID(),
		// 		ItemID:   drop.Resource.ID,
		// 		ItemType: "resources",
		// 		Quantity: drop.Quantity,
		// 	})
		// 	if err != nil {
		// 		return err
		// 	}
		// }

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
