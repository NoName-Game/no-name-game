package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func AbilityTree(update tgbotapi.Update, player models.Player) {
	message := update.Message
	routeName := "ability-tree"
	state := helpers.StartAndCreatePlayerState(routeName, player)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage", player.Language.Slug)
	switch state.Stage {
	case 0:
		if helpers.InStatsArray(message.Text) && player.Stats.AbilityPoint > 0 {
			state.Stage = 1
			validationFlag = true
		} else if player.Stats.AbilityPoint == 0 {
			state.Stage = 1
			validationMessage = "Punti abilità non sufficienti"
		}
	case 1:
		if message.Text == "Torna all'albero" {
			state.Stage = 0
			player.Update()
		} else if message.Text == "Esci" {
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
		text := "Quale statistica vuoi incrementare?\n"
		text += player.Stats.ToString()
		text += "\nHai a disposizione %d punti abilità!"
		msg := services.NewMessage(player.ChatID, fmt.Sprintf(text, player.Stats.AbilityPoint))
		msg.ReplyMarkup = helpers.StatsKeyboard()
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	case 1:
		if validationFlag {
			text := "Hai incrementato con successo " + message.Text + " !"
			player.Stats.Increment(message.Text)
			player.Update()
			msg := services.NewMessage(player.ChatID, text)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Torna all'albero"), tgbotapi.NewKeyboardButton("Esci")))
			services.SendMessage(msg)
		}
	case 2:
		if validationFlag {
			//====================================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, player)
			//====================================
			msg := services.NewMessage(player.ChatID, "Fine")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			services.SendMessage(msg)
		}
	}

}
