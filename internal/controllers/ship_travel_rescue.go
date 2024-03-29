package controllers

import (
	"fmt"
	"time"

	"nn-telegram/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/internal/helpers"
)

// ====================================
// ShipTravelRescueController
// ====================================
type ShipTravelRescueController struct {
	Payload struct {
		ItemID   uint32
		Quantity int32
	}
	Controller
}

func (c *ShipTravelRescueController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship.travel.rescue",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBlocked: []string{"exploration", "hunting"},
			ControllerBack: ControllerBack{
				To:        &ShipTravelController{},
				FromStage: 0,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
				2: {"route.breaker.clears", "route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipTravelRescueController) Handle(player *pb.Player, update tgbotapi.Update) {
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
func (c *ShipTravelRescueController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	case 0:
		// ##################################################################################################
		// Verifico che la nave equipaggiata non sia in riparazione
		// ##################################################################################################
		for _, state := range c.Data.PlayerActiveStates {
			if state.GetController() == "route.ship.travel.finding" {
				c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.controller.blocked")
				return true
			}
		}
	// ##################################################################################################
	// Verifico se il player ha richiesto il soccorso
	// ##################################################################################################
	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.travel.rescue.diamond") {
			return true
		}

	// ##################################################################################################
	// Verifico conferma player
	// ##################################################################################################
	case 2:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *ShipTravelRescueController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player se vuole esser trasportato al pianeta sicuro più vicino
	// ##################################################################################################
	case 0:
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "ship.travel.rescue.info"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "ship.travel.rescue.diamond")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1

	// ##################################################################################################
	// Chiedo conferma del trasferimento
	// ##################################################################################################
	case 1:
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "ship.travel.rescue.confirm"))

		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2

	// ##################################################################################################
	// Trasferminto confermato
	// ##################################################################################################
	case 2:
		if _, err := config.App.Server.Connection.TravelRescue(helpers.NewContext(1), &pb.TravelRescueRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			// Messaggio errore completamento
			msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "ship.travel.complete_diamond_error"))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
				),
			)

			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}

			// Fondamentale, esco senza chiudere
			c.CurrentState.Completed = true
			c.ForceBackTo = true
			return
		}

		// Invio messaggi di chiusura
		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "ship.travel.rescue.confirmed"))
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		var msgInit tgbotapi.Message
		if msgInit, err = helpers.SendMessage(helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "ship.travel.rescue.countdown.init"))); err != nil {
			c.Logger.Panic(err)
		}

		// Mando primo set di messaggi
		for i := 3; i >= 0; i-- {
			time.Sleep(2 * time.Second)
			edited := helpers.NewEditMessage(c.Player.ChatID, msgInit.MessageID,
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ship.travel.rescue.countdown"), i),
			)
			edited.ParseMode = tgbotapi.ModeHTML

			if _, err = helpers.SendMessage(edited); err != nil {
				c.Logger.Panic(err)
			}
		}

		// Forzo cancellazione posizione player in cache
		_ = helpers.DelPlayerPlanetPositionInCache(c.Player.GetID())

		// Completo lo stato
		c.CurrentState.Completed = true
		c.Configurations.ControllerBack.To = &MenuController{}
	}

	return
}
