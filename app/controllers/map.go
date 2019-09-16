package controllers

import (
	"encoding/json"
	"strings"
	"time"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/provider"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var mapKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", "map_up")),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", "map_left"), tgbotapi.NewInlineKeyboardButtonData("⭕", "map_action"), tgbotapi.NewInlineKeyboardButtonData("➡️", "map_right")),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", "map_down")),
)
var fightKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", "map_up")),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", "map_left"), tgbotapi.NewInlineKeyboardButtonData("⚔️", "map_fight.start"), tgbotapi.NewInlineKeyboardButtonData("➡️", "map_right")),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", "map_down")),
)

func MapController(update tgbotapi.Update) {
	callback := update.CallbackQuery

	// PAYLOAD
	type payloadStruct struct {
		Selection uint // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		InFight   bool
	}

	routeName := "callback.map"
	state := helpers.StartAndCreatePlayerState(routeName, helpers.Player)

	var payload payloadStruct
	helpers.UnmarshalPayload(state.Payload, &payload)

	m, _ := provider.GetMapByID(helpers.Player.ID)
	if m.ID < 1 {
		// Initialize payload
		payloadUpdated, _ := json.Marshal(payloadStruct{})
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)
		// Create map
		m, _ = provider.CreateMap(helpers.Player.ID)
		msg := services.NewMessage(helpers.Player.ChatID, helpers.TextDisplay(m))
		msg.ReplyMarkup = mapKeyboard
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	} else {
		if time.Since(m.UpdatedAt).Seconds() > 1.0 {

			if payload.InFight {
				Fight(update)
				return
			}

			var cellMap [66][66]bool
			actionCompleted := false
			err := json.Unmarshal([]byte(m.Cell), &cellMap)
			if err != nil {
				services.ErrorHandler("Error during unmarshal", err)
			}
			switch callback.Data {
			case "map_up":
				if !cellMap[m.PlayerX-1][m.PlayerY] {
					m.PlayerX--
					actionCompleted = true
				} else {
					m.PlayerX++
				}
			case "map_down":
				if !cellMap[m.PlayerX+1][m.PlayerY] {
					m.PlayerX++
					actionCompleted = true
				} else {
					m.PlayerX--
				}
			case "map_left":
				if !cellMap[m.PlayerX][m.PlayerY-1] {
					m.PlayerY--
					actionCompleted = true
				} else {
					m.PlayerY++
				}
			case "map_right":
				if !cellMap[m.PlayerX][m.PlayerY+1] {
					m.PlayerY++
					actionCompleted = true
				} else {
					m.PlayerY--
				}
			case "map_fight.start":
				payload.InFight = true
				payloadUpdated, _ := json.Marshal(payload)
				state.Payload = string(payloadUpdated)
				state, _ = provider.UpdatePlayerState(state)
				Fight(update)
			case "map_fight.end":
				helpers.FinishAndCompleteState(state, helpers.Player)
				services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("complete")))
				return
			}

			_, err = provider.UpdateMap(m)
			if err != nil {
				services.ErrorHandler("Error while updating map", err)
			}

			if actionCompleted {
				msg := services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.TextDisplay(m))
				if strings.Contains(helpers.TextDisplay(m), "*") {
					msg.ReplyMarkup = &fightKeyboard
				} else {
					msg.ReplyMarkup = &mapKeyboard
				}

				msg.ParseMode = "HTML"
				services.SendMessage(msg)
			}
			// Rimuove rotella di caricamento dal bottone
			services.AnswerCallbackQuery(services.NewAnswer(callback.ID, "", false))
		} else {
			answer := services.NewAnswer(callback.ID, "1 second delay", false)
			services.AnswerCallbackQuery(answer)
		}
	}
}

var mobKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔼", "map_fight.up")),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🗾", "map_fight.returnMap"), tgbotapi.NewInlineKeyboardButtonData("⚔", "map_fight.hit"), tgbotapi.NewInlineKeyboardButtonData("🔍", "map_fight.analize")),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", "map_fight.down")),
)

func Fight(update tgbotapi.Update) {

	callback := update.CallbackQuery

	// PAYLOAD
	type payloadStruct struct {
		Selection uint // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		InFight   bool
	}
	bodyParts := [4]string{"head", "chest", "gauntlets", "leg"}

	state := helpers.GetPlayerStateByFunction(helpers.Player, "callback.map")

	var payload payloadStruct
	helpers.UnmarshalPayload(state.Payload, &payload)

	var editMessage tgbotapi.EditMessageTextConfig
	m, _ := provider.GetMapByID(helpers.Player.ID)

	mob, _ := provider.GetEnemyByID(helpers.Player.ID)
	if mob.ID < 1 {
		mob, _ = provider.Spawn(nnsdk.Enemy{})
	}

	switch callback.Data {
	case "map_fight.start":
		editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("combat.card", mob.Name, mob.LifePoint, mob.LifeMax, helpers.Trans(bodyParts[payload.Selection])))
		editMessage.ReplyMarkup = &mobKeyboard
	case "map_fight.returnMap":
		payload.InFight = false
		payloadUpdated, _ := json.Marshal(payload)
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)
		editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.TextDisplay(m))
		editMessage.ParseMode = "HTML"
		editMessage.ReplyMarkup = &mapKeyboard
	case "map_fight.up":
		if payload.Selection > 0 {
			payload.Selection--
		} else {
			payload.Selection = 3
		}
		payloadUpdated, _ := json.Marshal(payload)
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)
	case "map_fight.down":
		if payload.Selection < 3 {
			payload.Selection++
		} else {
			payload.Selection = 0
		}
		payloadUpdated, _ := json.Marshal(payload)
		state.Payload = string(payloadUpdated)
		state, _ = provider.UpdatePlayerState(state)
	}

	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
		editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("combat.card", mob.Name, mob.LifePoint, mob.LifeMax, helpers.Trans(bodyParts[payload.Selection])))
		editMessage.ReplyMarkup = &mobKeyboard
	}

	services.SendMessage(editMessage)
	services.AnswerCallbackQuery(services.NewAnswer(callback.ID, "", false))
}
