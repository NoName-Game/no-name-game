package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//====================================
// AbilityController
//====================================
type AbilityController struct {
	BaseController
	Payload struct {}
}

//====================================
// Handle
//====================================
func (c *AbilityController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	var err error
	var isNewState bool
	c.RouteName, c.Update, c.Message = "route.abilityTree", update, update.Message

	// Check current state for this routes
	c.State, isNewState = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// It's first message
	if isNewState {
		c.Stage()
		return
	}

	// Go to validator
	if !c.Validator() {
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}

		// Ok! Run!
		c.Stage()
		return
	}

	// Validator goes errors
	validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
	services.SendMessage(validatorMsg)
	return
}

//====================================
// Validator
//====================================
func (c *AbilityController) Validator() (hasErrors bool) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch c.State.Stage {
	case 0:
		// Verifico se l'abilità passata esiste nelle abilità censite e se il player ha punti disponibili
		if helpers.InStatsStruct(c.Message.Text) && helpers.Player.Stats.AbilityPoint > 0 {
			c.State.Stage = 1
			return false
		} else if helpers.Player.Stats.AbilityPoint == 0 {
			c.State.Stage = 2
			return false
		}
	case 1:
		if c.Message.Text == helpers.Trans("ability.back") {
			c.State.Stage = 0
			return false
		} else if c.Message.Text == helpers.Trans("exit") {
			c.State.Stage = 2
			return false
		}
	}

	return true
}

//====================================
// Stage
//====================================
func (c *AbilityController) Stage() {
	var err error

	switch c.State.Stage {
	case 0:
		// Invio messaggio con recao stats
		messageSummaryPlayerStats := helpers.Trans("ability.stats.type", helpers.PlayerStatsToString(&helpers.Player.Stats))
		messagePlayerTotalPoint := helpers.Trans("ability.stats.total_point", helpers.Player.Stats.AbilityPoint)

		msg := services.NewMessage(helpers.Player.ChatID, messageSummaryPlayerStats+messagePlayerTotalPoint)
		msg.ReplyMarkup = helpers.StatsKeyboard()
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	case 1:
		// Invio Messaggio di incremento abilità
		text := helpers.Trans("ability.stats.completed", c.Message.Text)
		msg := services.NewMessage(helpers.Player.ChatID, text)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("ability.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("exit")),
			),
		)
		services.SendMessage(msg)

		// Incremento statistiche e aggiorno
		helpers.PlayerStatsIncrement(&helpers.Player.Stats, c.Message.Text)
		_, err = providers.UpdatePlayerStats(helpers.Player.Stats)
		if err != nil {
			services.ErrorHandler("Cant update player stats", err)
		}
	case 2:
		// Recap statistiche player
		text := helpers.Trans("ability.stats.type", helpers.PlayerStatsToString(&helpers.Player.Stats))
		if helpers.Player.Stats.AbilityPoint == 0 {
			text += "\n" + helpers.Trans("ability.no_point_left")
		} else {
			text += helpers.Trans("ability.stats.total_point", helpers.Player.Stats.AbilityPoint)
		}

		// Invio messaggio
		msg := services.NewMessage(helpers.Player.ChatID, text)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)

		// ====================================
		// COMPLETE!
		// ====================================
		helpers.FinishAndCompleteState(c.State, helpers.Player)
		// ====================================
	}
}
