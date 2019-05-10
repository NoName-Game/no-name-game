package controllers

import (
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/provider"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func AbilityTree(update tgbotapi.Update) {
	message := update.Message
	routeName := "route.abilityTree"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)

	playerStats, err := provider.GetPlayerStats(helpers.Player)
	if err != nil {
		services.ErrorHandler("Cant get player stats", err)
	}

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage")
	switch state.Stage {
	case 0:
		if helpers.InStatsStruct(message.Text) && playerStats.AbilityPoint > 0 {
			state.Stage = 1
			validationFlag = true
		} else if playerStats.AbilityPoint == 0 {
			state.Stage = 2
			validationFlag = true
		}
	case 1:
		if message.Text == helpers.Trans("ability.back") {
			state.Stage = 0
			state, _ = provider.UpdatePlayerState(state)
		} else if message.Text == helpers.Trans("exit") {
			state.Stage = 2
			validationFlag = true
		}
	}
	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			validatorMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			services.SendMessage(validatorMsg)
		}
	}

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		messageSummaryPlayerStats := helpers.Trans("ability.stats.type", helpers.PlayerStatsToString(&playerStats))
		messagePlayerTotalPoint := helpers.Trans("ability.stats.total_point", playerStats.AbilityPoint)

		msg := services.NewMessage(helpers.Player.ChatID, messageSummaryPlayerStats+messagePlayerTotalPoint)
		msg.ReplyMarkup = helpers.StatsKeyboard()
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	case 1:
		if validationFlag {
			// Increment player stats
			helpers.PlayerStatsIncrement(&playerStats, message.Text)

			playerStats, err = provider.UpdatePlayerStats(playerStats)
			if err != nil {
				services.ErrorHandler("Cant update player stats", err)
			}

			text := helpers.Trans("ability.stats.completed", message.Text)
			msg := services.NewMessage(helpers.Player.ChatID, text)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("ability.back")),
					tgbotapi.NewKeyboardButton(helpers.Trans("exit")),
				),
			)
			services.SendMessage(msg)
		}
	case 2:
		if validationFlag {
			// ====================================
			// IMPORTANT!
			// ====================================
			helpers.FinishAndCompleteState(state, helpers.Player)
			// ====================================

			text := helpers.Trans("ability.stats.type", helpers.PlayerStatsToString(&playerStats))
			if playerStats.AbilityPoint == 0 {
				text += "\n" + helpers.Trans("ability.no_point_left")
			} else {
				text += helpers.Trans("ability.stats.total_point", playerStats.AbilityPoint)
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

}
