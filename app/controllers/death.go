package controllers

import (
	"time"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/providers"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// TestTimedQuest - ...
func PlayerDeath(update tgbotapi.Update) {
	message := update.Message
	routeName := "route.death"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("playerDie", state.FinishAt.Format("3:04PM"))
	switch state.Stage {
	case 1:
		if time.Now().Before(state.FinishAt) {
			validationFlag = false
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			validatorMsg.ParseMode = "HTML"
			services.SendMessage(validatorMsg)
		}
	}

	switch state.Stage {
	case 0:
		// Set Timeout
		t := new(bool)
		*t = true

		state.Stage = 1
		state.ToNotify = t
		state.FinishAt = time.Now().Add((time.Hour * time.Duration(12)))
		state, _ = providers.UpdatePlayerState(state)

		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("playerDie", state.FinishAt.Format("3:04PM")))
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	case 1:
		if validationFlag {
			stats, err := providers.GetPlayerStats(helpers.Player)
			if err != nil {
				services.ErrorHandler("Cant retrieve stats", err)
			}
			*stats.LifePoint = 100 + stats.Level*10
			_, err = providers.UpdatePlayerStats(stats)
			if err != nil {
				services.ErrorHandler("Cant update stats", err)
			}
			helpers.FinishAndCompleteState(state, helpers.Player)
		}
	}

}
