package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type AbilityController struct {
	Update     tgbotapi.Update
	Message    *tgbotapi.Message
	RouteName  string
	Validation struct {
		HasErrors bool
		Message   string
	}
	Payload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}
	// Additional Data
	AddResourceFlag bool
}

//====================================
// Handle
//====================================
func (c *AbilityController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	c.RouteName = "route.abilityTree"
	c.Update = update
	c.Message = update.Message

	// Set Additional Data
	c.AddResourceFlag = false

	// Check current state for this routes
	state, isNewState := helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// It's first message
	if isNewState {
		c.Stage(state)
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(state.Payload, c.Payload)

	// Go to validator
	c.Validation.HasErrors, state = c.Validator(state)
	if !c.Validation.HasErrors {
		state, _ = providers.UpdatePlayerState(state)
		c.Stage(state)
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
func (c *AbilityController) Validator(state nnsdk.PlayerState) (hasErrors bool, newState nnsdk.PlayerState) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch state.Stage {
	case 0:
		if helpers.InStatsStruct(c.Message.Text) && helpers.Player.Stats.AbilityPoint > 0 {
			state.Stage = 1
			return false, state
		} else if helpers.Player.Stats.AbilityPoint == 0 {
			state.Stage = 2
			return false, state
		}
	case 1:
		if c.Message.Text == helpers.Trans("ability.back") {
			state.Stage = 0
			return false, state
		} else if c.Message.Text == helpers.Trans("exit") {
			state.Stage = 2
			return false, state
		}
	}

	return true, state
}

//====================================
// Stage
//====================================
func (c *AbilityController) Stage(state nnsdk.PlayerState) {
	switch state.Stage {
	case 0:
		messageSummaryPlayerStats := helpers.Trans("ability.stats.type", helpers.PlayerStatsToString(&helpers.Player.Stats))
		messagePlayerTotalPoint := helpers.Trans("ability.stats.total_point", helpers.Player.Stats.AbilityPoint)

		msg := services.NewMessage(helpers.Player.ChatID, messageSummaryPlayerStats+messagePlayerTotalPoint)
		msg.ReplyMarkup = helpers.StatsKeyboard()
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	case 1:
		// Increment player stats
		helpers.PlayerStatsIncrement(&helpers.Player.Stats, c.Message.Text)

		_, err := providers.UpdatePlayerStats(helpers.Player.Stats)
		if err != nil {
			services.ErrorHandler("Cant update player stats", err)
		}

		text := helpers.Trans("ability.stats.completed", c.Message.Text)
		msg := services.NewMessage(helpers.Player.ChatID, text)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("ability.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("exit")),
			),
		)
		services.SendMessage(msg)

	case 2:
		// ====================================
		// IMPORTANT!
		// ====================================
		helpers.FinishAndCompleteState(state, helpers.Player)
		// ====================================

		text := helpers.Trans("ability.stats.type", helpers.PlayerStatsToString(&helpers.Player.Stats))
		if helpers.Player.Stats.AbilityPoint == 0 {
			text += "\n" + helpers.Trans("ability.no_point_left")
		} else {
			text += helpers.Trans("ability.stats.total_point", helpers.Player.Stats.AbilityPoint)
		}

		msg := services.NewMessage(helpers.Player.ChatID, text)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)
	}
}
