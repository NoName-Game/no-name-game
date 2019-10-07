package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func Hunting(update tgbotapi.Update) {

	message := update.Message
	routeName := "route.hunting"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)

	// Stupid pointer stupid json pff
	t := new(bool)
	*t = true

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage")
	switch state.Stage {
	case 0:
		// Check if the player have a weapon equipped.
		if _, noWeapon := providers.GetPlayerWeapons(helpers.Player, "true"); noWeapon != nil {
			validationMessage = helpers.Trans("hunting.error.noWeaponEquipped")
			helpers.FinishAndCompleteState(state, helpers.Player)
		} else {
			validationFlag = true
		}
	case 1:
		if state.FinishAt.Before(time.Now()) {
			validationFlag = true
		} else {
			validationMessage = helpers.Trans("wait", state.FinishAt.Format("15:04:05"))
		}
	case 2:
		state := helpers.GetPlayerStateByFunction(helpers.Player, "callback.map")
		if state == (nnsdk.PlayerState{}) {
			validationFlag = true
		} else {
			MapController(update)
			return
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(helpers.Player.ChatID, validationMessage)
			services.SendMessage(validatorMsg)
		}
	}
	//====================================
	// LOGIC FLUX:
	// Waiting -> Map -> Drop -> Finish
	//====================================

	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		// Set timer
		state.FinishAt = helpers.GetEndTime(0, 10, 0)
		state.ToNotify = t
		state.Stage = 1
		_, err := providers.UpdatePlayerState(state)
		if err != nil {
			services.ErrorHandler("Cant update state", err)
		}
		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("hunting.searching", state.FinishAt.Format("04:05"))))
	case 1:
		if validationFlag {
			// Join Map
			state.Stage = 2
			state, _ = providers.UpdatePlayerState(state)
			MapController(update)
		}
	case 2:
		if validationFlag {
			//====================================
			// IMPORTANT!
			//====================================
			helpers.FinishAndCompleteState(state, helpers.Player)
			//====================================

			msg := services.NewMessage(message.Chat.ID, helpers.Trans("hunting.complete"))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				),
			)
			services.SendMessage(msg)
		}
	}
}
