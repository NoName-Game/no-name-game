package controllers

import (
	"encoding/json"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetTitanController
// ====================================
type SafePlanetTitanController struct {
	BaseController
	Payload struct{}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetTitanController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.safeplanet.titan",
		c.Payload,
		[]string{},
		player,
		update,
	) {
		return
	}

	// Verifico se vuole tornare indietro di stato
	if !proxy {
		if c.BackTo(1, &SafePlanetCoalitionController{}) {
			return
		}
	}

	// Set and load payload
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

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetTitanController) Validator() (hasErrors bool, err error) {
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

	case 1:
		// Recupero quali titani sono stati scoperti e quindi raggiungibili
		var rTitanDiscovered *pb.TitanDiscoveredResponse
		rTitanDiscovered, err = services.NnSDK.TitanDiscovered(helpers.NewContext(1), &pb.TitanDiscoveredRequest{})
		if err != nil {
			return
		}

		// Verifico sei il player ha passato il nome di un titano valido
		if len(rTitanDiscovered.GetTitans()) > 0 {
			for _, titan := range rTitanDiscovered.GetTitans() {
				if c.Update.Message.Text == titan.GetName() {
					return false, err
				}
			}
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true, err
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetTitanController) Stage() (err error) {
	switch c.CurrentState.Stage {
	case 0:
		var restsRecap string
		restsRecap = helpers.Trans(c.Player.Language.Slug, "route.safeplanet.titan.info")
		var keyboardRow [][]tgbotapi.KeyboardButton

		// Recupero quali titani sono stati scoperti e quindi raggiungibili
		var rTitanDiscovered *pb.TitanDiscoveredResponse
		rTitanDiscovered, err = services.NnSDK.TitanDiscovered(helpers.NewContext(1), &pb.TitanDiscoveredRequest{})
		if err != nil {
			return
		}

		// Se sono stati trovati dei tiani costruisco keyboard
		if len(rTitanDiscovered.GetTitans()) > 0 {
			restsRecap += helpers.Trans(c.Player.Language.Slug, "route.safeplanet.titan.choice")
			for _, titan := range rTitanDiscovered.GetTitans() {
				newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						// helpers.Trans(c.Player.Language.Slug, "ship.rests.start"),
						titan.GetName(),
					),
				)
				keyboardRow = append(keyboardRow, newKeyboardRow)
			}
		} else {
			restsRecap += helpers.Trans(c.Player.Language.Slug, "route.safeplanet.titan.no_titans_founded")
		}

		// Aggiungo torna indietro
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := services.NewMessage(c.Player.ChatID, restsRecap)
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
		c.CurrentState.Stage = 1

	// In questo stage avvio effettivamente il riposo
	case 1:
		// Recupero pianeta da titano
		var rGetTitanByName *pb.GetTitanByNameResponse
		rGetTitanByName, err = services.NnSDK.GetTitanByName(helpers.NewContext(1), &pb.GetTitanByNameRequest{
			Name: c.Update.Message.Text,
		})
		if err != nil {
			return
		}

		// Aggiunto nuova posizione al player
		_, err = services.NnSDK.CreatePlayerPosition(helpers.NewContext(1), &pb.CreatePlayerPositionRequest{
			PlayerID: c.Player.ID,
			PlanetID: rGetTitanByName.GetTitan().GetPlanetID(),
		})
		if err != nil {
			return
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "route.safeplanet.titan.teleport"),
		)

		msg.ParseMode = "markdown"
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
