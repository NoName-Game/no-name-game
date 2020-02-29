package controllers

import (
	"encoding/json"
	"math"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipRestsController
// ====================================
type ShipRestsController struct {
	BaseController
	Payload struct {
		StartDateTime time.Time
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipRestsController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	var playerStateProvider providers.PlayerStateProvider

	c.Controller = "route.ship.rests"
	c.Player = player
	c.Update = update

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// Validate
	var hasError bool
	hasError, err = c.Validator()
	if err != nil {
		panic(err)
	}

	// Se ritornano degli errori
	if hasError == true {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
		validatorMsg.ParseMode = "markdown"
		// validatorMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		// 	tgbotapi.NewKeyboardButtonRow(
		// 		tgbotapi.NewKeyboardButton(
		// 			helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
		// 		),
		// 	),
		// )

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
	c.State.Payload = string(payloadUpdated)
	c.State, err = playerStateProvider.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}

	return
}

// ====================================
// Validator
// ====================================
func (c *ShipRestsController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	switch c.State.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.rests.start") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			return true, err
		}

		return false, err
	case 2:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.rests.validator.need_to_wakeup")
			return true, err
		}

		return false, err
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *ShipRestsController) Stage() (err error) {
	var playerProvider providers.PlayerProvider

	switch c.State.Stage {

	// In questo riporto al player le tempistiche necesarie al riposo
	case 0:
		// Recupero informazioni per il recupero totale delle energie
		var restsInfo nnsdk.PlayerRestInfoResponse
		restsInfo, err = playerProvider.GetRestsInfo(c.Player.ID)
		if err != nil {
			return err
		}

		// Costruisco info per riposo
		var restsRecap string
		restsRecap = helpers.Trans(c.Player.Language.Slug, "ship.rests.info")
		if restsInfo.NeedRests {
			restsRecap += helpers.Trans(c.Player.Language.Slug, "ship.rests.time", restsInfo.RestsTime)
		} else {
			restsRecap = helpers.Trans(c.Player.Language.Slug, "ship.rests.dont_need")
		}

		// Aggiongo bottone start riposo
		var keyboardRow [][]tgbotapi.KeyboardButton
		if restsInfo.NeedRests {
			newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "ship.rests.start"),
				),
			)
			keyboardRow = append(keyboardRow, newKeyboardRow)
		}

		// Aggiungo abbandona solo se il player non è morto e quindi obbligato a dormire
		if *c.Player.Stats.Dead == false {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
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
		c.State.Stage = 1

	// In questo stage avvio effettivamente il riposo
	case 1:
		// Recupero informazioni per il recupero totale delle energie
		var restsInfo nnsdk.PlayerRestInfoResponse
		restsInfo, err = playerProvider.GetRestsInfo(c.Player.ID)
		if err != nil {
			return err
		}

		// Setto timer recuperato dalla chiamata delle info
		c.State.FinishAt = helpers.GetEndTime(0, int(restsInfo.RestsTime), 0)

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.rests.reparing", c.State.FinishAt.Format("15:04:05")),
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

		// Aggiorno stato
		c.Payload.StartDateTime = time.Now()
		c.State.Stage = 2
	case 2:
		// Calcolo differenza tra inizio e fine
		var endDate time.Time
		endDate = time.Now()

		diffDate := endDate.Sub(c.Payload.StartDateTime)
		diffMinutes := math.RoundToEven(diffDate.Minutes())

		// Fine riparazione
		err = playerProvider.EndPlayerRest(c.Player.ID, nnsdk.PlayerRestEndRequest{
			RestsTime: uint(diffMinutes),
		})
		if err != nil {
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.rests.reparing.finish"),
		)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
				),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.State.Completed = helpers.SetTrue()
	}

	return
}
