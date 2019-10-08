package controllers

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type HuntingController struct {
	Update     tgbotapi.Update
	Message    *tgbotapi.Message
	Callback   *tgbotapi.CallbackQuery
	RouteName  string
	Validation struct {
		HasErrors bool
		Message   string
	}
	Payload        struct{}
	HuntingPayload struct {
		Selection uint // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		InFight   bool
		Kill      uint
	}
	// Additional Data
	State nnsdk.PlayerState
}

//====================================
// Handle
//====================================
func (c *HuntingController) Handle(update tgbotapi.Update) {
	// Check message type
	if update.Message != nil {
		// Current Controller instance
		c.RouteName = "route.hunting"
		c.Update = update
		c.Message = update.Message
		c.State, _ = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

		// Set and load payload
		helpers.UnmarshalPayload(c.State.Payload, c.Payload)

		// Go to validator
		c.Validation.HasErrors, c.State = c.Validator(c.State)
		if !c.Validation.HasErrors {
			c.State, _ = providers.UpdatePlayerState(c.State)
			c.Stage(c.State)
			return
		}

		// Validator goes errors
		validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
		services.SendMessage(validatorMsg)
		return

	} else if update.CallbackQuery != nil {
		// Current Controller instance
		c.RouteName = "callback.map"
		c.Update = update
		c.Callback = update.CallbackQuery
		c.State, _ = helpers.CheckState(c.RouteName, c.HuntingPayload, helpers.Player)

		c.hunting()
	}

}

//====================================
// Validator
//====================================
func (c *HuntingController) Validator(state nnsdk.PlayerState) (hasErrors bool, newState nnsdk.PlayerState) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch state.Stage {
	case 0:
		// Check if the player have a weapon equipped.
		if _, noWeapon := providers.GetPlayerWeapons(helpers.Player, "true"); noWeapon != nil {
			c.Validation.Message = helpers.Trans("hunting.error.noWeaponEquipped")
			helpers.FinishAndCompleteState(state, helpers.Player)
			return true, state
		}
		return false, state
	case 1:
		c.Validation.Message = helpers.Trans("wait", state.FinishAt.Format("15:04:05"))
		if state.FinishAt.Before(time.Now()) {
			return false, state
		}
	case 2:
		state, _ = helpers.GetPlayerStateByFunction(helpers.Player, "callback.map")
		if state == (nnsdk.PlayerState{}) {
			return false, state
		}

		// else {
		// 	MapController(update)
		// 	// new(MenuController).Handle(update)
		// 	return
		// }
	}

	return true, state
}

//====================================
// Stage Waiting -> Map -> Drop -> Finish
//====================================
func (c *HuntingController) Stage(state nnsdk.PlayerState) {
	switch state.Stage {
	case 0:
		// Set timer
		state.FinishAt = helpers.GetEndTime(0, 10, 0)
		t := new(bool)
		*t = true
		state.ToNotify = t
		state.Stage = 1
		_, err := providers.UpdatePlayerState(state)
		if err != nil {
			services.ErrorHandler("Cant update state", err)
		}

		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("hunting.searching", state.FinishAt.Format("04:05"))))
	case 1:
		// Join Map
		state.Stage = 2
		state, _ = providers.UpdatePlayerState(state)
		// MapController(update)
		c.hunting()
	case 2:
		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(state, helpers.Player)
		//====================================

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("hunting.complete"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)

	}
}

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

func (c *HuntingController) hunting() {
	m, _ := providers.GetMapByID(helpers.Player.ID)
	if m.ID < 1 {
		// Create map
		m, _ = providers.CreateMap(helpers.Player.ID)
		msg := services.NewMessage(helpers.Player.ChatID, helpers.TextDisplay(m))
		msg.ReplyMarkup = mapKeyboard
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	} else {
		// Verificare Perchè dovrebbe arrivare un messaggio normale

		// if update.Message != nil {
		// msg := services.NewMessage(helpers.Player.ChatID, helpers.TextDisplay(m))
		// if strings.Contains(helpers.TextDisplay(m), "*") {
		// 	msg.ReplyMarkup = &fightKeyboard
		// } else {
		// 	msg.ReplyMarkup = &mapKeyboard
		// }
		// msg.ParseMode = "HTML"
		// services.SendMessage(msg)
		// return
		// }
		if time.Since(m.UpdatedAt).Seconds() > 1.0 {

			if c.HuntingPayload.InFight {
				//Fight(update)
				log.Panicln("Implementami")
				return
			}

			var cellMap [66][66]bool
			actionCompleted := false
			err := json.Unmarshal([]byte(m.Cell), &cellMap)
			if err != nil {
				services.ErrorHandler("Error during unmarshal", err)
			}
			switch c.Callback.Data {
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
				c.HuntingPayload.InFight = true
				payloadUpdated, _ := json.Marshal(c.HuntingPayload)
				c.State.Payload = string(payloadUpdated)
				c.State, _ = providers.UpdatePlayerState(c.State)
				// Fight(update)
				log.Panicln("Implementami")

			case "map_finish.fight":
				helpers.FinishAndCompleteState(c.State, helpers.Player)
				services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, c.Callback.Message.MessageID, helpers.Trans("complete")))
				return
			case "map_no.action":
				actionCompleted = true
			}

			_, err = providers.UpdateMap(m)
			if err != nil {
				services.ErrorHandler("Error while updating map", err)
			}

			if actionCompleted {
				msg := services.NewEditMessage(helpers.Player.ChatID, c.Callback.Message.MessageID, helpers.TextDisplay(m))
				if strings.Contains(helpers.TextDisplay(m), "*") {
					msg.ReplyMarkup = &fightKeyboard
				} else {
					msg.ReplyMarkup = &mapKeyboard
				}

				msg.ParseMode = "HTML"
				services.SendMessage(msg)
			}
			// Rimuove rotella di caricamento dal bottone
			services.AnswerCallbackQuery(services.NewAnswer(c.Callback.ID, "", false))
		} else {
			answer := services.NewAnswer(c.Callback.ID, "1 second delay", false)
			services.AnswerCallbackQuery(answer)
		}
	}
}
