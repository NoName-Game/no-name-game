package controllers

import (
	"encoding/json"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// TitanController
// ====================================
type TitanController struct {
	BaseController
	Payload struct{}
}

// ====================================
// Handle
// ====================================
func (c *TitanController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
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
		if c.BackTo(1, &CoalitionController{}) {
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
func (c *TitanController) Validator() (hasErrors bool, err error) {
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
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.rests.start") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			return true, err
		}

		return false, err
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *TitanController) Stage() (err error) {
	switch c.CurrentState.Stage {

	// In questo riporto al player le tempistiche necesarie al riposo
	case 0:
		// Costruisco info per riposo
		var restsRecap string
		restsRecap = helpers.Trans(c.Player.Language.Slug, "route.safeplanet.titan.info")

		// TODO: Aggiunto bottoni su quali titani sono stati scoperti e raggiungibili

		var keyboardRow [][]tgbotapi.KeyboardButton
		// if restsInfo.NeedRests {
		// 	newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
		// 		tgbotapi.NewKeyboardButton(
		// 			helpers.Trans(c.Player.Language.Slug, "ship.rests.start"),
		// 		),
		// 	)
		// 	keyboardRow = append(keyboardRow, newKeyboardRow)
		// }

		// Aggiungo abbandona solo se il player non è morto e quindi obbligato a dormire
		if !c.PlayerStats.GetDead() {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			))
		}

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
		//TODO: Verifico verso quale pianeta/titano è stato scelto di teletrasportsi

		// Ripulisco messaggio per recupermi solo il nome
		// equipmentName = strings.Split(c.Update.Message.Text, " (")[0]

		//TODO: Chiamo WS che aggiorneà la posizione del player

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.rests.reparing"),
		)

		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup")),
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
