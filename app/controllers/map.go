package controllers

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

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
		Kill      uint
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
			case "map_finish.fight":
				helpers.FinishAndCompleteState(state, helpers.Player)
				services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("complete")))
				return
			case "map_no.action":
				actionCompleted = true
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
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🗾", "map_fight.returnMap"), tgbotapi.NewInlineKeyboardButtonData("⚔", "map_fight.hit")),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", "map_fight.down")),
)

func Fight(update tgbotapi.Update) {

	callback := update.CallbackQuery

	// PAYLOAD
	type payloadStruct struct {
		Selection uint // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		InFight   bool
		Kill      uint
	}
	bodyParts := [4]string{"head", "chest", "gauntlets", "leg"}

	state := helpers.GetPlayerStateByFunction(helpers.Player, "callback.map")

	var payload payloadStruct
	helpers.UnmarshalPayload(state.Payload, &payload)

	var editMessage tgbotapi.EditMessageTextConfig
	m, _ := provider.GetMapByID(helpers.Player.ID)

	mob := m.Enemies[helpers.ChooseMob(m)]

	switch callback.Data {
	case "map_fight.start":
		// editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("combat.card", mob.Name, mob.LifePoint, mob.LifeMax, helpers.Trans(bodyParts[payload.Selection])))
		// editMessage.ReplyMarkup = &mobKeyboard
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
	case "map_fight.hit":
		mobDistance, _ := provider.Distance(m, mob)
		mobPercentage := ((1000 - mobDistance) / 1000) // What percentage I see of the body? Number between 0 -> 1
		//var damageMultiplier float64
		precision, _ := provider.PlayerPrecision(helpers.Player.ID, payload.Selection)
		precision *= (85.0 / 37.0) * mobPercentage // Base precision

		if rand.Float64() < precision {
			// Hitted
			_, err := provider.GetPlayerWeapons(helpers.Player, "true")
			if err != nil {
				services.ErrorHandler("Error while retriving weapons", err)
			}
			playerDamage, _ := provider.PlayerDamage(helpers.Player.ID)
			damageToMob := uint(playerDamage)
			mob.LifePoint -= damageToMob
			if mob.LifePoint > mob.LifeMax || mob.LifePoint == 0 {
				// Mob die
				payload.Kill++
				mob.LifePoint = 0
				_, err := provider.DeleteEnemy(mob.ID)
				if err != nil {
					services.ErrorHandler("Cant delete enemy.", err)
				}
				editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("combat.mob_killed"))
				var ok tgbotapi.InlineKeyboardMarkup
				if payload.Kill == uint(len(m.Enemies)) {
					ok = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "map_finish.fight")))
				} else {
					ok = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "map_no.action")))
				}
				editMessage.ReplyMarkup = &ok
				// Add drop
				stats, _ := provider.GetPlayerStats(helpers.Player)
				helpers.IncrementExp(1, stats)
				payload.InFight = false
				payloadUpdated, _ := json.Marshal(payload)
				state.Payload = string(payloadUpdated)
				state, _ = provider.UpdatePlayerState(state)
			} else {
				damageToPlayer, _ := provider.EnemyDamage(mob.ID)
				stats, _ := provider.GetPlayerStats(helpers.Player)
				stats = helpers.DecrementLife(uint(damageToPlayer), stats)
				if *stats.LifePoint == 0 {
					// Player Die
					helpers.DeleteRedisAndDbState(helpers.Player)
					msg := services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("playerDie"))
					msg.ParseMode = "HTML"
					services.SendMessage(msg)
					helpers.FinishAndCompleteState(state, helpers.Player)
				} else {
					editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("combat.damage", damageToMob, uint(damageToPlayer)))
					ok := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "map_no.action")))
					editMessage.ReplyMarkup = &ok
				}
				_, err = provider.UpdateEnemy(mob)
				if err != nil {
					services.ErrorHandler("Error while updating enemy", err)
				}
			}
		} else {
			// Miss by player
			damageToPlayer, _ := provider.EnemyDamage(mob.ID)
			stats, _ := provider.GetPlayerStats(helpers.Player)
			stats = helpers.DecrementLife(uint(damageToPlayer), stats)
			if *stats.LifePoint == 0 {
				// Player Die
				helpers.DeleteRedisAndDbState(helpers.Player)
				msg := services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("playerDie"))
				msg.ParseMode = "HTML"
				services.SendMessage(msg)
				helpers.FinishAndCompleteState(state, helpers.Player)
			} else {
				editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("combat.miss", damageToPlayer))
				ok := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "map_no.action")))
				editMessage.ReplyMarkup = &ok
			}
		}

	}

	// Standard Message
	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
		stats, _ := provider.GetPlayerStats(helpers.Player)
		editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("combat.card", mob.Name, mob.LifePoint, mob.LifeMax, helpers.Player.Username, &stats.LifePoint, (100+stats.Level*10), helpers.Trans(bodyParts[payload.Selection])))
		editMessage.ReplyMarkup = &mobKeyboard
	}

	services.SendMessage(editMessage)
	services.AnswerCallbackQuery(services.NewAnswer(callback.ID, "", false))
}
