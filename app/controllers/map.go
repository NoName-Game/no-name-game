package controllers

import (
	"encoding/json"
	"math"
	"math/rand"
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
			case "map_finish.fight":
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
		// alt(B)=1,7*[1000 - dist(B)]/1000
		mobDistance := math.Sqrt(math.Pow(float64(m.EnemyX-m.PlayerX), 2) + math.Pow(float64(m.EnemyY-m.PlayerY), 2))
		// 1000 corrisponde ad 1KM, la distanza in cui si vede circa lo 0% del corpo, con la formula calcoliamo quanta percentuale del corpo vedo
		// per determinare la percentuale di riuscita di nel colpire una parte del corpo.
		mobPercentage := ((1000 - mobDistance) / 1000) // What percentage I see of the body? Number between 0 -> 1
		// RULE OF NINE: HEAD > 9%, BODY > 37%, ARMS > 18, LEGS > 36
		// 1 : 90 = mobPercentage : x
		var damageMultiplier float64
		var precision float64
		precision = (85.0 / 37.0) * mobPercentage // Base precision
		switch payload.Selection {
		case 0: // HEAD
			precision *= 9.0
			damageMultiplier = 3
		case 1: // BODY
			precision *= 37.0 // precision * body part weight
			damageMultiplier = 1
		case 2: // ARMS
			precision *= 18.0
			damageMultiplier = 0.9
		case 3: // LEGS
			precision *= 36.0
			damageMultiplier = 0.9
		}
		if rand.Float64() < precision {
			// Hitted
			playerWeapons, err := provider.GetPlayerWeapons(helpers.Player, "true")
			if err != nil {
				services.ErrorHandler("Error while retriving weapons", err)
			}
			damageToMob := uint(math.Round(damageMultiplier * 1.2 * ((rand.Float64() * 10) + float64(playerWeapons[0].RawDamage))))
			mob.LifePoint -= damageToMob
			if mob.LifePoint > mob.LifeMax || mob.LifePoint == 0 {
				// Mob die
				mob.LifePoint = 0
				editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("combat.mob_killed"))
				ok := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "map_finish.fight")))
				editMessage.ReplyMarkup = &ok
				// Add drop
				playerStats, _ := provider.GetPlayerStats(helpers.Player)
				playerStats.Experience++
				provider.UpdatePlayerStats(playerStats)
				payload.InFight = false
				payloadUpdated, _ := json.Marshal(payload)
				state.Payload = string(payloadUpdated)
				state, _ = provider.UpdatePlayerState(state)
			} else {
				damageToPlayer := uint(math.Round(damageMultiplier * 1.2 * ((rand.Float64() * 10) + 4)))
				stats, _ := provider.GetPlayerStats(helpers.Player)
				helpers.DecrementLife(damageToPlayer, stats)
				editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("combat.damage", damageToMob, damageToPlayer))
				ok := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "map_no.action")))
				editMessage.ReplyMarkup = &ok
			}
			_, err = provider.UpdateEnemy(mob)
			if err != nil {
				services.ErrorHandler("Error while updating enemy", err)
			}
		}

	}

	// Standard Message
	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
		editMessage = services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, helpers.Trans("combat.card", mob.Name, mob.LifePoint, mob.LifeMax, helpers.Trans(bodyParts[payload.Selection])))
		editMessage.ReplyMarkup = &mobKeyboard
	}

	services.SendMessage(editMessage)
	services.AnswerCallbackQuery(services.NewAnswer(callback.ID, "", false))
}
