package controllers

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
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
func (c *ShipRestsController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.ship.rests",
		c.Payload,
		[]string{"hunting", "mission"},
		player,
		update,
	) {
		return
	}

	// Verifico se vuole tornare indietro di stato
	if !proxy {
		if c.BackTo(1, &ShipController{}) {
			return
		}
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
	c.State.Payload = string(payloadUpdated)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := services.NnSDK.UpdatePlayerState(ctx, &pb.UpdatePlayerStateRequest{
		PlayerState: c.State,
	})
	if err != nil {
		panic(err)
	}
	c.State = response.GetPlayerState()

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *ShipRestsController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		),
	)

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

		// Si vuole svegliare, ma non è passato ancora un minuto
		var endDate = time.Now()

		diffDate := endDate.Sub(c.Payload.StartDateTime)
		diffMinutes := math.RoundToEven(diffDate.Minutes())
		if diffMinutes <= 1 {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.rests.need_to_rest")
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
	switch c.State.Stage {

	// In questo riporto al player le tempistiche necesarie al riposo
	case 0:

		// Recupero informazioni per il recupero totale delle energie
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		restsInfo, err := services.NnSDK.GetRestsInfo(ctx, &pb.GetRestsInfoRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			panic(err)
		}

		// Costruisco info per riposo
		var restsRecap string
		restsRecap = helpers.Trans(c.Player.Language.Slug, "ship.rests.info")
		if restsInfo.NeedRests {
			restsRecap += helpers.Trans(c.Player.Language.Slug, "ship.rests.time", restsInfo.GetRestsTime())
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
		if !c.Player.GetStats().GetDead() {
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
		c.State.Stage = 1

	// In questo stage avvio effettivamente il riposo
	case 1:
		// Recupero informazioni per il recupero totale delle energie
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		restsInfo, err := services.NnSDK.GetRestsInfo(ctx, &pb.GetRestsInfoRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			panic(err)
		}

		// Setto timer recuperato dalla chiamata delle info
		finishTime := helpers.GetEndTime(0, int(restsInfo.GetRestsTime()), 0)
		c.State.FinishAt, _ = ptypes.TimestampProto(finishTime)

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.rests.reparing", finishTime.Format("15:04:05")),
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := services.NnSDK.EndPlayerRest(ctx, &pb.EndPlayerRestRequest{
			PlayerID:  c.Player.GetID(),
			RestsTime: uint32(diffMinutes),
		})
		if err != nil {
			panic(err)
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.rests.reparing.finish", diffMinutes),
		)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
				),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.State.Completed = true
	}

	return
}
