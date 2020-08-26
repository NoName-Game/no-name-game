package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
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
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetExpansionController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller: "route.safeplanet.coalition.expansion",
		ControllerBack: ControllerBack{
			To:        &SafePlanetCoalitionController{},
			FromStage: 2,
		},
		Payload: c.Payload,
	}) {
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)

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
	c.PlayerData.CurrentState.Payload = string(payloadUpdated)

	rUpdatePlayerState, err := services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
		PlayerState: c.PlayerData.CurrentState,
	})
	if err != nil {
		panic(err)
	}
	c.PlayerData.CurrentState = rUpdatePlayerState.GetPlayerState()

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetExpansionController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		),
	)

	switch c.PlayerData.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	case 1:
		// Verifico se il nome passato è quello di un pianeta sicuro
		var rGetSafePlanets *pb.GetSafePlanetsResponse
		rGetSafePlanets, err = services.NnSDK.GetSafePlanets(helpers.NewContext(1), &pb.GetSafePlanetsRequest{})
		if err != nil {
			return
		}

		planetName := strings.Split(c.Update.Message.Text, " -")[0]

		// Verifico sei il player ha passato il nome di un titano valido
		if len(rGetSafePlanets.GetSafePlanets()) > 0 {
			for _, planet := range rGetSafePlanets.GetSafePlanets() {
				if planetName == planet.GetName() {
					return false, err
				}
			}
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true, err
		// Verifico la conferma dell'uso
	case 2:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "confirm") {
			// Verifico che il player abbia anche abbastanza soldi
			var rGetPlayerEconomy *pb.GetPlayerEconomyResponse
			rGetPlayerEconomy, err = services.NnSDK.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
				PlayerID:    c.Player.ID,
				EconomyType: "diamond",
			})
			if err != nil {
				return
			}

			if rGetPlayerEconomy.GetValue() >= int32(c.Payload.Price) {
				return false, err
			}

			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.teleport_ko")
			return true, err
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true, err
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetExpansionController) Stage() (err error) {
	switch c.PlayerData.CurrentState.Stage {
	case 0:
		var expansionRecap string
		expansionRecap = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.info")
		var keyboardRow [][]tgbotapi.KeyboardButton

		// Recupero quanti pianeti mancano per l'ampliamento del sistema
		var rGetExpansionInfo *pb.GetExpansionInfoResponse
		rGetExpansionInfo, err = services.NnSDK.GetExpansionInfo(helpers.NewContext(1), &pb.GetExpansionInfoRequest{})
		if err != nil {
			return
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
		rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			panic(err)
		}

		// Mostro la lista dei pianeti sicuri disponibili
		var rGetSafePlanets *pb.GetTeletrasportSafePlanetListResponse
		rGetSafePlanets, err = services.NnSDK.GetTeletrasportSafePlanetList(helpers.NewContext(1), &pb.GetTeletrasportSafePlanetListRequest{
			PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
		})
		if err != nil {
			return
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
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := services.NewMessage(c.Player.ChatID, expansionRecap)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.PlayerData.CurrentState.Stage = 1

	case 1:
		// Recupero quale pianeta vole raggiungere e dettaglio costo
		planetNameChoiched := strings.Split(c.Update.Message.Text, " -")[0]

		// Recupero ultima posizione del player
		var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
		rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			panic(err)
		}

		var rGetSafePlanets *pb.GetTeletrasportSafePlanetListResponse
		rGetSafePlanets, err = services.NnSDK.GetTeletrasportSafePlanetList(helpers.NewContext(1), &pb.GetTeletrasportSafePlanetListRequest{
			PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
		})
		if err != nil {
			return
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
		msg := services.NewMessage(c.Update.Message.Chat.ID,
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

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.Payload.PlanetID = planet.GetID()
		c.Payload.Price = int32(price)
		c.PlayerData.CurrentState.Stage = 2
	case 2:
		// Concludo teletrasporto
		_, err = services.NnSDK.EndTeletrasportSafePlanet(helpers.NewContext(1), &pb.EndTeletrasportSafePlanetRequest{
			PlayerID: c.Player.ID,
			PlanetID: c.Payload.PlanetID,
			Price:    -c.Payload.Price,
		})
		if err != nil {
			return
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.expansion.teleport_ok"),
		)

		msg.ParseMode = "markdown"
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.PlayerData.CurrentState.Completed = true
		c.Configuration.ControllerBack.To = &MenuController{}
	}

	return
}
