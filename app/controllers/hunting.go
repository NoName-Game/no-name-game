package controllers

import (
	"encoding/json"
	"time"

	"bitbucket.org/no-name-game/no-name/app/commands"
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func Hunting(update tgbotapi.Update, player models.Player) {

	message := update.Message
	routeName := "hunting"
	state := helpers.StartAndCreatePlayerState(routeName, player)

	type payloadHunting struct {
		Mob   models.Enemy
		Score int //Number of enemy defeated
	}

	var payload payloadHunting
	helpers.UnmarshalPayload(state.Payload, &payload)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage", player.Language.Slug)
	switch state.Stage {
	case 1:
		if state.FinishAt.Before(time.Now()) {
			validationFlag = true
		} else {
			validationMessage = helpers.Trans("wait", player.Language.Slug, state.FinishAt.Format("15:04:05"))
		}
	case 2:
		if message.Text == helpers.Trans("continue", player.Language.Slug) {
			validationFlag = true
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(player.ChatID, validationMessage)
			services.SendMessage(validatorMsg)
		}
	}

	//====================================
	// LOGIC FLUX:
	// Searching -> Finding -> Fight [Loop] -> Drop
	//====================================

	// FIGHT SYSTEM : Enemy Card / Choose Attack -> Calculate Damage -> Apply Damage

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		// Set timer
		state.FinishAt = commands.GetEndTime(0, 5, 0)
		state.ToNotify = true
		state.Stage = 1

		//FIXME: Pick a random enemy from database
		payload.Mob = models.Enemy{
			Name:             "Spider",
			Life:             1000,
			DamageMultiplier: 1.0,
		}
		payload.Score = 1
		payloadUpdated, _ := json.Marshal(payload)
		state.Payload = string(payloadUpdated)
		state.Update()

		services.SendMessage(services.NewMessage(player.ChatID, helpers.Trans("searching", player.Language.Slug, state.FinishAt.Format("15:04:05"))))
	case 1:
		if validationFlag {
			// Enemy found
			state.Stage = 2
			state.Update()
			msg := services.NewMessage(player.ChatID, helpers.Trans("enemy_found", player.Language.Slug, payload.Mob.Name))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("continue", player.Language.Slug))))
			services.SendMessage(msg)
		}
	case 2:
		if validationFlag {
			//START BATTLE
			// Scheme
			// Name
			// Life
			// ExtraEffect (Optional)
			// Cosa vuoi fare?
			msg := services.NewMessage(player.ChatID, helpers.Trans("enemy_card", player.Language.Slug, payload.Mob.Name, payload.Mob.Life))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Attacca con Arma1")))

		}
	}
}
