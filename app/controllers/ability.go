package controllers

import (
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func AbilityTree(update tgbotapi.Update, player models.Player) {
	message := update.Message
	routeName := "route.abilityTree"
	state := helpers.StartAndCreatePlayerState(routeName, player)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage", player.Language.Slug)
	switch state.Stage {
	case 0:
		if helpers.InStatsStruct(message.Text, player.Language.Slug) && player.Stats.AbilityPoint > 0 {
			state.Stage = 1
			validationFlag = true
		} else if player.Stats.AbilityPoint == 0 {
			state.Stage = 1
			validationMessage = helpers.Trans("ability.no_point_left", player.Language.Slug)
		}
	case 1:
		if message.Text == helpers.Trans("ability.back", player.Language.Slug) {
			state.Stage = 0
			player.Update()
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
		messageSummaryPlayerStats := helpers.Trans("ability.stats.type", player.Language.Slug, player.Stats.ToString(player.Language.Slug))
		messagePlayerTotalPoint := helpers.Trans("ability.stats.total_point", player.Language.Slug, player.Stats.AbilityPoint)

		msg := services.NewMessage(player.ChatID, messageSummaryPlayerStats+messagePlayerTotalPoint)
		msg.ReplyMarkup = helpers.StatsKeyboard(player.Language.Slug)
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	case 1:
		if validationFlag {
			// Increment player stast
			player.Stats.Increment(message.Text, player.Language.Slug)
			player.Update()

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
			//====================================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, player)
			//====================================

			//TODO: Call main menu

			msg := services.NewMessage(player.ChatID, helpers.Trans("complete", player.Language.Slug))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back", player.Language.Slug)),
				),
			)
			services.SendMessage(msg)
		}
	}

}
