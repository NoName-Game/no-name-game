package controllers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/provider"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func AbilityTree(update tgbotapi.Update, player nnsdk.Player) {
	message := update.Message
	routeName := "route.abilityTree"
	state := helpers.StartAndCreatePlayerState(routeName, player)

	playerStats, err := provider.GetPlayerStats(player)
	if err != nil {
		services.ErrorHandler("Cant get player stats", err)
	}

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage", player.Language.Slug)
	switch state.Stage {
	case 0:
		if helpers.InStatsStruct(message.Text, player.Language.Slug) && playerStats.AbilityPoint > 0 {
			state.Stage = 1
			validationFlag = true
		} else if playerStats.AbilityPoint == 0 {
			state.Stage = 2
			validationFlag = true
		}
	case 1:
		if message.Text == helpers.Trans("ability.back", player.Language.Slug) {
			state.Stage = 0
			state, _ = provider.UpdatePlayerState(state)
		} else if message.Text == helpers.Trans("exit", player.Language.Slug) {
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
		messageSummaryPlayerStats := helpers.Trans("ability.stats.type", player.Language.Slug, helpers.PlayerStatsToString(&playerStats, player.Language.Slug))
		messagePlayerTotalPoint := helpers.Trans("ability.stats.total_point", player.Language.Slug, playerStats.AbilityPoint)

		msg := services.NewMessage(player.ChatID, messageSummaryPlayerStats+messagePlayerTotalPoint)
		msg.ReplyMarkup = helpers.StatsKeyboard(player.Language.Slug)
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	case 1:
		if validationFlag {
			// Increment player stats
			// player.Stats.Increment(message.Text, player.Language.Slug)

			helpers.PlayerStatsIncrement(&playerStats, message.Text, player.Language.Slug)

			playerStats, err = provider.UpdatePlayerStats(playerStats)
			if err != nil {
				services.ErrorHandler("Cant update player stats", err)
			}

			text := helpers.Trans("ability.stats.completed", player.Language.Slug, message.Text)
			msg := services.NewMessage(player.ChatID, text)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("ability.back", player.Language.Slug)),
					tgbotapi.NewKeyboardButton(helpers.Trans("exit", player.Language.Slug)),
				),
			)
			services.SendMessage(msg)
		}
	case 2:
		if validationFlag {
			// ====================================
			// IMPORTANT!
			// ====================================
			helpers.FinishAndCompleteState(state, player)
			// ====================================

			text := helpers.Trans("ability.stats.type", player.Language.Slug, helpers.PlayerStatsToString(&playerStats, player.Language.Slug))
			if playerStats.AbilityPoint == 0 {
				text += "\n" + helpers.Trans("ability.no_point_left", player.Language.Slug)
			} else {
				text += helpers.Trans("ability.stats.total_point", player.Language.Slug, playerStats.AbilityPoint)
			}

			msg := services.NewMessage(player.ChatID, text)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
				),
			)
			services.SendMessage(msg)
		}
	}

}
