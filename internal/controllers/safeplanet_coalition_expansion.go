package controllers

import (
	"fmt"
	"strings"

	"nn-grpc/build/pb"
	"nn-telegram/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-telegram/internal/helpers"
)

// ====================================
// SafePlanetExpansionController
// ====================================
type SafePlanetExpansionController struct {
	Payload struct {
		PlanetID uint32
		Price    int32
	}
	Controller
}

func (c *SafePlanetExpansionController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.expansion",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 1,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				1: {"route.breaker.menu"},
				2: {"route.breaker.back"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetExpansionController) Handle(player *pb.Player, update tgbotapi.Update) {
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
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetExpansionController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il nome passato è quello di un pianeta sicuro
	// ##################################################################################################
	case 1:
		var err error
		var rGetSafePlanets *pb.GetSafePlanetsResponse
		if rGetSafePlanets, err = config.App.Server.Connection.GetSafePlanets(helpers.NewContext(1), &pb.GetSafePlanetsRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		planetName := strings.Split(c.Update.Message.Text, " -")[0]

		// Verifico sei il player ha passato il nome di un titano valido
		if len(rGetSafePlanets.GetSafePlanets()) > 0 {
			for _, planet := range rGetSafePlanets.GetSafePlanets() {
				if planetName == planet.GetName() {
					return false
				}
			}
		}

		return true

	// ##################################################################################################
	// Verifico la conferma dell'uso
	// ##################################################################################################
	case 2:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "confirm") {
			var err error
			var rGetPlayerEconomy *pb.GetPlayerEconomyResponse
			if rGetPlayerEconomy, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
				PlayerID:    c.Player.ID,
				EconomyType: pb.GetPlayerEconomyRequest_DIAMOND,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Verifico che il player ha abbastanza soldi
			if rGetPlayerEconomy.GetValue() >= c.Payload.Price {
				return false
			}

			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.teleport_ko")
			return true
		}

		return true
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetExpansionController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		var expansionRecap string
		expansionRecap = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.info")
		rank := helpers.Trans(c.Player.Language.Slug, "rank."+c.Player.Rank.NameCode)
		expansionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.player_rank", rank, c.Player.RankID)
		var keyboardRow [][]tgbotapi.KeyboardButton

		// Recupero quanti pianeti mancano per l'ampliamento del sistema
		var rGetExpansionInfo *pb.GetExpansionInfoResponse
		if rGetExpansionInfo, err = config.App.Server.Connection.GetExpansionInfo(helpers.NewContext(1), &pb.GetExpansionInfoRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		// Messaggio titanto non ancora sconfitto
		var rGetTitanByPlanetSystemID *pb.GetTitanByPlanetSystemIDResponse
		if rGetTitanByPlanetSystemID, err = config.App.Server.Connection.GetTitanByPlanetSystemID(helpers.NewContext(1), &pb.GetTitanByPlanetSystemIDRequest{
			PlanetSystemID: rGetExpansionInfo.GetLastSystemDiscovered().GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// #######################
		// Recap messaggi
		// #######################

		// Messaggio ultimo sistema scoperto
		expansionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.last_system",
			rGetExpansionInfo.GetLastSystemDiscovered().GetName(),
		)

		// Messaggio pianeti che mancano da esplorare
		if rGetExpansionInfo.GetMissPlanetsCounter() <= 0 && rGetTitanByPlanetSystemID.GetTitan().GetDied() {
			expansionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.done")
		} else {
			// Recap pianeti
			if rGetExpansionInfo.GetMissPlanetsCounter() > 0 {
				expansionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.recap_planets",
					rGetExpansionInfo.GetMissPlanetsCounter(),
					rGetExpansionInfo.GetTotalPlanetsCounter(),
				)
			} else {
				expansionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.recap_planets_done")
			}

			// Recap titano
			if rGetTitanByPlanetSystemID.GetTitan().GetDied() {
				expansionRecap += helpers.Trans(c.Player.Language.Slug,
					"safeplanet.coalition.expansion.recap_titan_done",
					rGetTitanByPlanetSystemID.GetTitan().GetName(),
				)
			} else {
				expansionRecap += helpers.Trans(c.Player.Language.Slug,
					"safeplanet.coalition.expansion.recap_titan",
					rGetTitanByPlanetSystemID.GetTitan().GetName(),
				)
			}
		}

		// #######################
		// Lista pianeti sicuri raggiungibili
		// #######################

		// Recupero posizione player corrente
		var playerPosition *pb.Planet
		if playerPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
			c.Logger.Panic(err)
		}

		// Mostro la lista dei pianeti sicuri disponibili
		var rGetSafePlanets *pb.GetTeletrasportSafePlanetListResponse
		if rGetSafePlanets, err = config.App.Server.Connection.GetTeletrasportSafePlanetList(helpers.NewContext(1), &pb.GetTeletrasportSafePlanetListRequest{
			PlanetID: playerPosition.GetID(),
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		if len(rGetSafePlanets.GetSafePlanets()) > 0 {
			expansionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.choice")
			for _, safePlanet := range rGetSafePlanets.GetSafePlanets() {
				newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						fmt.Sprintf("%s - 💎%v", safePlanet.GetPlanet().GetName(), safePlanet.GetPrice()),
					),
				)
				keyboardRow = append(keyboardRow, newKeyboardRow)
			}
		}

		// Aggiungo torna indietro
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.Player.ChatID, expansionRecap)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	case 1:
		// Recupero quale pianeta vole raggiungere e dettaglio costo
		planetNameChoiched := strings.Split(c.Update.Message.Text, " -")[0]

		// Recupero posizione player corrente
		var playerPosition *pb.Planet
		if playerPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
			c.Logger.Panic(err)
		}

		var rGetSafePlanets *pb.GetTeletrasportSafePlanetListResponse
		if rGetSafePlanets, err = config.App.Server.Connection.GetTeletrasportSafePlanetList(helpers.NewContext(1), &pb.GetTeletrasportSafePlanetListRequest{
			PlanetID: playerPosition.GetID(),
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero costo finale
		var planet *pb.Planet
		var price uint32
		for _, safePlanet := range rGetSafePlanets.GetSafePlanets() {
			if safePlanet.GetPlanet().GetName() == planetNameChoiched {
				planet = safePlanet.GetPlanet()
				price = safePlanet.GetPrice()
			}
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.confirmation",
				planet.GetName(),
				price,
			),
		)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.Payload.PlanetID = planet.GetID()
		c.Payload.Price = int32(price)
		c.CurrentState.Stage = 2
	case 2:
		// Concludo teletrasporto
		if _, err = config.App.Server.Connection.EndTeletrasportSafePlanet(helpers.NewContext(1), &pb.EndTeletrasportSafePlanetRequest{
			PlayerID: c.Player.ID,
			PlanetID: c.Payload.PlanetID,
			Price:    -c.Payload.Price,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Cancello posizione su cache così che venga rigenerata correttamente. NNWS-62
		if err = helpers.DelPlayerPlanetPositionInCache(c.Player.ID); err != nil {
			c.Logger.Warning(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.teleport_ok"),
		)

		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
		c.Configurations.ControllerBack.To = &MenuController{}
	}

	return
}
