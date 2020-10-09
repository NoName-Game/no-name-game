package controllers

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ExplorationController
// ====================================
type ExplorationController struct {
	Controller
	ExplorationTypeChoiched string // Esplorazione scelta dall'utente
}

// ====================================
// Handle
// ====================================
func (c *ExplorationController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.exploration",
		},
		Configurations: ControllerConfigurations{
			CustomBreaker:     []string{"exploration.breaker.continue"},
			ControllerBlocked: []string{"hunting", "ship"},
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 1,
			},
		},
	}) {
		return
	}

	// Validate
	if c.Validator() {
		c.Validate()
		return
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(nil)
}

// ====================================
// Validator
// ====================================
func (c *ExplorationController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il player ha passato una tipoligia di esplorazione valida
	// ##################################################################################################
	case 1:
		var rGetAllExplorationCategories *pb.GetAllExplorationCategoriesResponse
		if rGetAllExplorationCategories, err = config.App.Server.Connection.GetAllExplorationCategories(helpers.NewContext(1), &pb.GetAllExplorationCategoriesRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		// Controllo se il messaggio continene uno dei tipi di missione dichiarati
		for _, missionType := range rGetAllExplorationCategories.GetExplorationCategories() {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("exploration.%s", missionType.GetSlug())) == c.Update.Message.Text {
				c.ExplorationTypeChoiched = missionType.GetSlug()
				return false
			}
		}

		return true
	// ##################################################################################################
	// Verifica stato esplorazione
	// ##################################################################################################
	case 2:
		var rExplorationCheck *pb.ExplorationCheckResponse
		if rExplorationCheck, err = config.App.Server.Connection.ExplorationCheck(helpers.NewContext(1), &pb.ExplorationCheckRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Il Player deve terminare prima l'esplorazione in corso
		if !rExplorationCheck.GetFinishExploration() {
			var finishAt time.Time
			if finishAt, err = helpers.GetEndTime(rExplorationCheck.GetExplorationEndTime(), c.Player); err != nil {
				c.Logger.Panic(err)
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
						helpers.Trans(c.Player.Language.Slug, "exploration.breaker.continue"),
					),
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
					),
				),
			)

			return true
		}
	// ##################################################################################################
	// Verifico se il player vuole continuare o terminare l'esplorazione
	// ##################################################################################################
	case 3:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "exploration.continue") {
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "exploration.comeback") {
			c.CurrentState.Stage = 4
			return false
		}

		return true
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *ExplorationController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// Primo avvio di missione, restituisco al player
	// i vari tipi di missioni disponibili
	case 0:
		// Recupero tutte le categorie di esplorazione possibili
		var rGetAllExplorationCategories *pb.GetAllExplorationCategoriesResponse
		if rGetAllExplorationCategories, err = config.App.Server.Connection.GetAllExplorationCategories(helpers.NewContext(1), &pb.GetAllExplorationCategoriesRequest{}); err != nil {
			c.Logger.Panic(err)
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
		if rGetPlayerCurrentPlanet, err = config.App.Server.Connection.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Verifico se sono conquistatore
		var rGetCurrentConquerorByPlanetID *pb.GetCurrentConquerorByPlanetIDResponse
		if rGetCurrentConquerorByPlanetID, err = config.App.Server.Connection.GetCurrentConquerorByPlanetID(helpers.NewContext(1), &pb.GetCurrentConquerorByPlanetIDRequest{
			PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Verifico se il player è un conquistatore
		if c.Player.ID == rGetCurrentConquerorByPlanetID.GetPlayer().GetID() {
			messageExploration += helpers.Trans(c.Player.Language.Slug, "exploration.conqueror_bonus")
		}

		// Invio messaggi con il tipo di missioni come tastierino
		msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s\n\n%s",
			messageExploration,
			helpers.Trans(c.Player.Language.Slug, "exploration.tips"),
		))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRows,
			ResizeKeyboard: true,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1

	// In questo stage verrà recuperato il tempo di attesa per il
	// completamnto della missione e notificato al player
	case 1:
		// Avvio nuova esplorazione
		var rExplorationStart *pb.ExplorationStartResponse
		if rExplorationStart, err = config.App.Server.Connection.ExplorationStart(helpers.NewContext(1), &pb.ExplorationStartRequest{
			PlayerID:                c.Player.ID,
			ExplorationCategorySlug: c.ExplorationTypeChoiched,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Converto finishAt in formato Time
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rExplorationStart.GetFinishAt(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio di attesa
		msg := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"exploration.wait",
				finishAt.Format("15:04:05"),
			),
		)
		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stage
		c.CurrentState.Stage = 2
		c.ForceBackTo = true

	// In questo stage recupero quali risorse il player ha recuperato
	// dalla missione e glielo notifico
	case 2:
		var rExplorationCheck *pb.ExplorationCheckResponse
		if rExplorationCheck, err = config.App.Server.Connection.ExplorationCheck(helpers.NewContext(1), &pb.ExplorationCheckRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero dattigli risorsa
		var rGetResourceByID *pb.GetResourceByIDResponse
		if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
			ID: rExplorationCheck.GetDropResult().GetResourceID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio di riepilogo con le materie recuperate e chiedo se vuole continuare o ritornare
		msg := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"exploration.extraction_recap",
				rGetResourceByID.GetResource().GetName(),
				rGetResourceByID.GetResource().GetRarity().GetName(),
				strings.ToUpper(rGetResourceByID.GetResource().GetRarity().GetSlug()),
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

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno lo stato
		c.CurrentState.Stage = 3

	// In questo stage verifico cosa ha scelto di fare il player
	// se ha deciso di continuare allora ritornerò ad uno stato precedente,
	// mentre se ha deciso di concludere andrò avanti di stato
	case 3:
		// Continua esplorazione
		var rExplorationContinue *pb.ExplorationContinueResponse
		if rExplorationContinue, err = config.App.Server.Connection.ExplorationContinue(helpers.NewContext(1), &pb.ExplorationContinueRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Converto finishAt in formato Time
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rExplorationContinue.GetFinishAt(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		// Il player ha scelto di continuare la ricerca
		msg := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"exploration.wait",
				finishAt.Format("15:04:05"),
			),
		)
		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno lo stato
		c.CurrentState.Stage = 2
		c.ForceBackTo = true

	// Ritorno il messaggio con gli elementi droppati
	case 4:
		var rExplorationEnd *pb.ExplorationEndResponse
		if rExplorationEnd, err = config.App.Server.Connection.ExplorationEnd(helpers.NewContext(1), &pb.ExplorationEndRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recap delle risorse ricavate da questa missione
		var dropList string
		for _, drop := range rExplorationEnd.GetDropResults() {
			// Recupero dattigli risorsa
			var rGetResourceByID *pb.GetResourceByIDResponse
			if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: drop.ResourceID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			dropList += fmt.Sprintf(
				"- %v x *%s* (%s)\n",
				drop.Quantity,
				rGetResourceByID.GetResource().GetName(),
				strings.ToUpper(rGetResourceByID.GetResource().GetRarity().GetSlug()),
			)
		}

		// Invio messaggio di chiusura missione
		msg := helpers.NewMessage(c.Player.ChatID,
			fmt.Sprintf("%s%s",
				helpers.Trans(c.Player.Language.Slug, "exploration.extraction_ended"),
				dropList,
			),
		)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
