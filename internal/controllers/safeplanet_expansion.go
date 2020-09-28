package controllers

import (
	"fmt"
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

// ====================================
// Handle
// ====================================
func (c *SafePlanetExpansionController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
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
		var keyboardRow [][]tgbotapi.KeyboardButton

		// Recupero quanti pianeti mancano per l'ampliamento del sistema
		var rGetExpansionInfo *pb.GetExpansionInfoResponse
		if rGetExpansionInfo, err = config.App.Server.Connection.GetExpansionInfo(helpers.NewContext(1), &pb.GetExpansionInfoRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		expansionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.last_system",
			rGetExpansionInfo.GetLastSystemDiscovered().GetName(),
		)

		if rGetExpansionInfo.GetMissPlanetsCounter() <= 0 {
			expansionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.done")
		} else {
			expansionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.recap",
				rGetExpansionInfo.GetMissPlanetsCounter(),
				rGetExpansionInfo.GetTotalPlanetsCounter(),
			)
		}

		// Recupero ultima posizione del player
		var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
		if rGetPlayerCurrentPlanet, err = config.App.Server.Connection.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Mostro la lista dei pianeti sicuri disponibili
		var rGetSafePlanets *pb.GetTeletrasportSafePlanetListResponse
		if rGetSafePlanets, err = config.App.Server.Connection.GetTeletrasportSafePlanetList(helpers.NewContext(1), &pb.GetTeletrasportSafePlanetListRequest{
			PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
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
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.Player.ChatID, expansionRecap)
		msg.ParseMode = "markdown"
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

		// Recupero ultima posizione del player
		var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
		if rGetPlayerCurrentPlanet, err = config.App.Server.Connection.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		var rGetSafePlanets *pb.GetTeletrasportSafePlanetListResponse
		if rGetSafePlanets, err = config.App.Server.Connection.GetTeletrasportSafePlanetList(helpers.NewContext(1), &pb.GetTeletrasportSafePlanetListRequest{
			PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
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
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.confirmation",
				planet.GetName(),
				price,
			),
		)
		msg.ParseMode = "markdown"
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

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.teleport_ok"),
		)

		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
		c.Configurations.ControllerBack.To = &MenuController{}
	}

	return
}
