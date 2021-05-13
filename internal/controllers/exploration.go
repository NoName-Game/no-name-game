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

func (c *ExplorationController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
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
			PlanetType: []string{"default", "titan"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.clears"},
				1: {"route.breaker.menu"},
				2: {"exploration.breaker.continue"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *ExplorationController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(c.Configuration(player, update)) {
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
	case 0:
		// ##################################################################################################
		// Verifico che sul pianeta non ci sia un titano
		// ##################################################################################################
		if inTitanPlanet, _ := c.CheckInTitanPlanet(c.Data.PlayerCurrentPosition); inTitanPlanet {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.titan_in_planet")
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
					),
				),
			)
			return true
		}

		// Verifico che il player non sia in viaggio
		if inTravel, _ := c.CheckInTravel(); inTravel {
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
					),
				),
			)
			return true
		}

	case 1:
		// ##################################################################################################
		// Verifico se il player ha passato una tipoligia di esplorazione valida
		// ##################################################################################################
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
			// Verificio se il player vuole forzare il ritorno alla base
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "exploration.cancel") {
				c.CurrentState.Stage = 4
				return false
			}

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
						helpers.Trans(c.Player.Language.Slug, "exploration.cancel"),
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
		var categories []*pb.ExplorationCategory
		if categories, err = helpers.GetExplorationCategories(); err != nil {
			c.Logger.Panic(err)
		}

		// Creo messaggio con la lista delle missioni possibili
		var keyboardRows [][]tgbotapi.KeyboardButton
		for _, missionType := range categories {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("exploration.%s", missionType.GetSlug()))),
			)

			keyboardRows = append(keyboardRows, keyboardRow)
		}

		// Aggiungo anche abbandona
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.menu"),
			),
		))

		// Costruisco risposta
		var messageExploration string
		messageExploration = helpers.Trans(c.Player.Language.Slug, "exploration.exploration")

		// Messaggio bonus
		messageExploration += c.CheckBonus()

		// Invio messaggi con il tipo di missioni come tastierino
		msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s\n\n%s",
			messageExploration,
			helpers.Trans(c.Player.Language.Slug, "exploration.tips"),
		))
		msg.ParseMode = tgbotapi.ModeHTML
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
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stage
		c.CurrentState.Stage = 2
		c.ForceBackTo = true

	// In questo stage recupero quali risorse il player ha recuperato
	// dalla missione e glielo notifico
	case 2:
		var rExplorationDropResources *pb.ExplorationDropResourcesResponse
		if rExplorationDropResources, err = config.App.Server.Connection.ExplorationDropResources(helpers.NewContext(1), &pb.ExplorationDropResourcesRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Riporto le risorese estratte in questo ciclo
		var cycleResourcesMessage string
		for _, dropResult := range rExplorationDropResources.GetCycleDropResults() {
			// Recupero dettagli risorse
			var rGetResourceByID *pb.GetResourceByIDResponse
			if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: dropResult.GetResourceID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Verifico se se la risorsa ha bonus
			var bonus string
			if dropResult.GetBonus() {
				bonus = helpers.Trans(c.Player.Language.Slug, "exploration.extraction_resource_duplicated")
			}

			// Aggiungo dettaglio risorsa
			cycleResourcesMessage += fmt.Sprintf("%s <b>%v</b> x <b>%s</b> (%s) %s %s\n",
				helpers.GetResourceCategoryIcons(rGetResourceByID.GetResource().GetResourceCategoryID()),
				dropResult.GetQuantity(),
				rGetResourceByID.GetResource().GetName(),
				rGetResourceByID.GetResource().GetRarity().GetSlug(),
				helpers.GetResourceBaseIcons(rGetResourceByID.GetResource().GetBase()),
				bonus,
			)
		}

		// TUTTE le risorse estratte
		var allResourcesMessage string
		for _, dropResult := range rExplorationDropResources.GetAllDropResults() {
			// Recupero dettagli risorse
			var rGetResourceByID *pb.GetResourceByIDResponse
			if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: dropResult.GetResourceID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Aggiungo dettaglio risorsa
			allResourcesMessage += fmt.Sprintf("%s <b>%v</b> x <b>%s</b> (%s) %s\n",
				helpers.GetResourceCategoryIcons(rGetResourceByID.GetResource().GetResourceCategoryID()),
				dropResult.GetQuantity(),
				rGetResourceByID.GetResource().GetName(),
				rGetResourceByID.GetResource().GetRarity().GetSlug(),
				helpers.GetResourceBaseIcons(rGetResourceByID.GetResource().GetBase()),
			)
		}

		// Invio messaggio di riepilogo con le materie recuperate e chiedo se vuole continuare o ritornare
		msg := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"exploration.extraction_recap",
				cycleResourcesMessage,
				allResourcesMessage,
			),
		)

		msg.ParseMode = tgbotapi.ModeHTML
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
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "exploration.continue")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "exploration.comeback")),
			),
		)

		// Aggiorno lo stato
		c.CurrentState.Stage = 2
		c.ForceBackTo = true

	// Ritorno il messaggio con gli elementi droppati
	case 4:
		var quickEnd bool
		quickEnd = false
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "exploration.cancel") {
			// Player ha annullato prima l'esplorazione
			quickEnd = true
		}
		var rExplorationEnd *pb.ExplorationEndResponse
		if rExplorationEnd, err = config.App.Server.Connection.ExplorationEnd(helpers.NewContext(1), &pb.ExplorationEndRequest{
			PlayerID: c.Player.ID,
			QuickEnd: quickEnd,
		}); err != nil {
			if strings.Contains(err.Error(), "inventory full") {
				msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "inventory.inventory_full"))
				msg.ParseMode = tgbotapi.ModeHTML
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				if _, err = helpers.SendMessage(msg); err != nil {
					c.Logger.Panic(err)
				}
				c.CurrentState.Completed = true
				return
			} else {
				c.Logger.Panic(err)
			}
		}

		var msg tgbotapi.MessageConfig
		if len(rExplorationEnd.GetDropResults()) == 0 {
			msg = helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "exploration.extraction_ended_zero"))
		} else {
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
					"- %s <b>%v</b> x <b>%s</b> (%s) %s\n",
					helpers.GetResourceCategoryIcons(rGetResourceByID.GetResource().GetResourceCategoryID()),
					drop.Quantity,
					rGetResourceByID.GetResource().GetName(),
					strings.ToUpper(rGetResourceByID.GetResource().GetRarity().GetSlug()),
					helpers.GetResourceBaseIcons(rGetResourceByID.GetResource().GetBase()),
				)
			}
			// Invio messaggio di chiusura missione
			msg = helpers.NewMessage(c.Player.ChatID,
				helpers.Trans(c.Player.Language.Slug, "exploration.extraction_ended",
					c.Data.PlayerCurrentPosition.Name,
					dropList,
					rExplorationEnd.GetExp(),
				),
			)
		}

		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}

// CheckBonus - Metodo di verifica bonus
func (c *ExplorationController) CheckBonus() (bonunsMessage string) {
	var err error

	// Recupero posizione player corrente
	var playerPosition *pb.Planet
	if playerPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
		c.Logger.Panic(err)
	}

	// Verifico se sono conquistatore
	var rGetCurrentConquerorByPlanetID *pb.GetCurrentConquerorByPlanetIDResponse
	if rGetCurrentConquerorByPlanetID, err = config.App.Server.Connection.GetCurrentConquerorByPlanetID(helpers.NewContext(1), &pb.GetCurrentConquerorByPlanetIDRequest{
		PlanetID: playerPosition.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Verifico se il player è un conquistatore
	if c.Player.ID == rGetCurrentConquerorByPlanetID.GetPlayer().GetID() {
		bonunsMessage += helpers.Trans(c.Player.Language.Slug, "exploration.conqueror_bonus")
	}

	// Verifico se il player ha bonus dominio
	var rGetCurrentDomainByPlanetID *pb.GetCurrentDomainByPlanetIDResponse
	if rGetCurrentDomainByPlanetID, err = config.App.Server.Connection.GetCurrentDomainByPlanetID(helpers.NewContext(1), &pb.GetCurrentDomainByPlanetIDRequest{
		PlanetID: playerPosition.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Recupero gilda player
	var rGetPlayerGuild *pb.GetPlayerGuildResponse
	if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	if rGetPlayerGuild.GetGuild().GetID() == rGetCurrentDomainByPlanetID.GetGuild().GetID() {
		bonunsMessage += helpers.Trans(c.Player.Language.Slug, "exploration.domain_bonus")
	}

	return
}
