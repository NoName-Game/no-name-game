package controllers

import (
	"bitbucket.org/no-name-game/no-name/app/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StartMission - start an exploration
func StartMission(update tgbotapi.Update, player models.Player) {
	// routeName := "route.mission"
	// message := update.Message

	// type payloadStruct struct {
	// 	ExplorationType string
	// 	Times           int
	// 	Material        models.Resource
	// 	Quantity        int
	// }

	// state := helpers.StartAndCreatePlayerState(routeName, player)
	// var payload payloadStruct
	// helpers.UnmarshalPayload(state.Payload, &payload)

	// eTypes := make([]string, 3)
	// eTypes[0] = helpers.Trans("mission.underground", player.Language.Slug)
	// eTypes[1] = helpers.Trans("mission.surface", player.Language.Slug)
	// eTypes[2] = helpers.Trans("mission.atmosphere", player.Language.Slug)

	// //====================================
	// // Validator
	// //====================================
	// validationFlag := false
	// validationMessage := helpers.Trans("validationMessage", player.Language.Slug)
	// switch state.Stage {
	// case 1:
	// 	input := message.Text
	// 	if helpers.StringInSlice(input, eTypes) {
	// 		payload.ExplorationType = input
	// 		validationFlag = true
	// 	}
	// case 2:
	// 	if time.Now().After(state.FinishAt) && payload.Times < 10 {
	// 		payload.Times++
	// 		payload.Quantity = rand.Intn(3)*payload.Times + 1
	// 		validationFlag = true
	// 		validationMessage = helpers.Trans("mission.wait", player.Language.Slug, state.FinishAt.Format("2006-01-02 15:04:05"))
	// 	}
	// case 3:
	// 	input := message.Text
	// 	if input == helpers.Trans("mission.continue", player.Language.Slug) {
	// 		validationFlag = true
	// 		state.FinishAt = commands.GetEndTime(0, 10*(2*payload.Times), 0)
	// 		state.ToNotify = true
	// 	} else if input == helpers.Trans("mission.comeback", player.Language.Slug) {
	// 		state.Stage = 4
	// 		validationFlag = true
	// 	}
	// }

	// if !validationFlag {
	// 	if state.Stage != 0 {
	// 		validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
	// 		validatorMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	// 		services.SendMessage(validatorMsg)
	// 	}
	// }
	// //====================================
	// // Stage
	// //====================================
	// switch state.Stage {
	// case 0:
	// 	payloadUpdated, _ := json.Marshal(payloadStruct{})
	// 	state.Payload = string(payloadUpdated)
	// 	state.Stage = 1
	// 	state.Update()

	// 	msg := services.NewMessage(player.ChatID, helpers.Trans("mission.exploration", player.Language.Slug))

	// 	var keyboardRows [][]tgbotapi.KeyboardButton
	// 	for _, eType := range eTypes {
	// 		keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(eType))
	// 		keyboardRows = append(keyboardRows, keyboardRow)
	// 	}
	// 	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
	// 		Keyboard:       keyboardRows,
	// 		ResizeKeyboard: true,
	// 	}
	// 	services.SendMessage(msg)
	// case 1:
	// 	if validationFlag {
	// 		// FIXME: replace me with new method
	// 		payload.Material = models.GetRandomResourceByCategory(helpers.GetMissionCategoryID(message.Text, player.Language.Slug))

	// 		// Updating state
	// 		payloadUpdated, _ := json.Marshal(payload)
	// 		state.Payload = string(payloadUpdated)
	// 		state.Stage = 2
	// 		state.ToNotify = true

	// 		// Set finishAt
	// 		state.FinishAt = commands.GetEndTime(0, 10, 0)
	// 		state.Update()

	// 		// Remove current redist stare
	// 		helpers.DelRedisState(player)

	// 		msg := services.NewMessage(player.ChatID, helpers.Trans("mission.wait", player.Language.Slug, string(state.FinishAt.Format("15:04:05"))))
	// 		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	// 		services.SendMessage(msg)
	// 	}
	// case 2:
	// 	if validationFlag {
	// 		payloadUpdated, _ := json.Marshal(payload)
	// 		state.Payload = string(payloadUpdated)
	// 		state.Stage = 3
	// 		state.Update()

	// 		msg := services.NewMessage(player.ChatID, helpers.Trans("mission.extraction_recap", player.Language.Slug, payload.Material.Name, payload.Quantity))
	// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
	// 			tgbotapi.NewKeyboardButtonRow(
	// 				tgbotapi.NewKeyboardButton(helpers.Trans("mission.continue", player.Language.Slug)),
	// 				tgbotapi.NewKeyboardButton(helpers.Trans("mission.comeback", player.Language.Slug)),
	// 			),
	// 		)
	// 		services.SendMessage(msg)
	// 	}
	// case 3:
	// 	if validationFlag {
	// 		state.Stage = 2

	// 		// Remove current redist stare
	// 		helpers.DelRedisState(player)

	// 		// Continue
	// 		msg := services.NewMessage(player.ChatID, helpers.Trans("mission.wait", player.Language.Slug, string(state.FinishAt.Format("15:04:05"))))
	// 		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	// 		services.SendMessage(msg)
	// 		state.Update()
	// 	}
	// case 4:
	// 	if validationFlag {
	// 		// Exit
	// 		helpers.FinishAndCompleteState(state, player)

	// 		// Add Items to player inventory
	// 		player.Inventory.AddResource(payload.Material, payload.Quantity)

	// 		msg := services.NewMessage(player.ChatID, helpers.Trans("mission.extraction_ended", player.Language.Slug))
	// 		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	// 		services.SendMessage(msg)
	// 	}
	// }
}
