package controllers

import (
	"encoding/json"
	"log"
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

func MapController(update tgbotapi.Update) {
	callback := update.CallbackQuery

	m, _ := provider.GetMapByID(helpers.Player.ID)
	if m.ID < 1 {
		// Create map
		m, _ = provider.CreateMap()
		msg := services.NewMessage(helpers.Player.ChatID, "<code>"+helpers.TextDisplay(m)+"</code>")
		msg.ReplyMarkup = mapKeyboard
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	} else {
		if time.Since(m.UpdatedAt).Seconds() > 1.0 {

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
				}
			case "map_down":
				if !cellMap[m.PlayerX+1][m.PlayerY] {
					m.PlayerX++
					actionCompleted = true
				}
			case "map_left":
				if !cellMap[m.PlayerX][m.PlayerY-1] {
					m.PlayerY--
					actionCompleted = true
				}
			case "map_right":
				if !cellMap[m.PlayerX][m.PlayerY+1] {
					m.PlayerY++
					actionCompleted = true
				}
			}

			_, err = provider.UpdateMap(m)
			if err != nil {
				services.ErrorHandler("Error while updating map", err)
			}

			if actionCompleted {
				log.Println(helpers.Player.ChatID)
				msg := services.NewEditMessage(helpers.Player.ChatID, callback.Message.MessageID, "<code>"+helpers.TextDisplay(m)+"</code>")
				msg.ReplyMarkup = &mapKeyboard
				msg.ParseMode = "HTML"
				services.SendMessage(msg)
				services.AnswerCallbackQuery(services.NewAnswer(callback.ID, "", false))
			}
		} else {
			answer := services.NewAnswer(callback.ID, "1 second delay", false)
			services.AnswerCallbackQuery(answer)
		}
	}

}
