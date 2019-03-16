package controllers

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"bitbucket.org/no-name-game/no-name/app/commands"

	"bitbucket.org/no-name-game/no-name/services"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StartMission - start an exploration
func StartMission(update tgbotapi.Update, player models.Player) {

	routeName := "missione"
	message := update.Message

	type payloadStruct struct {
		ExplorationType string
		Times           int
		Material        models.Item
		Quantity        int
	}

	state := helpers.StartAndCreatePlayerState(routeName, player)
	var payload payloadStruct
	helpers.UnmarshalPayload(state.Payload, &payload)

	eTypes := make([]string, 3)
	eTypes[0] = helpers.Trans("sottosuolo", player.Language.Slug)
	eTypes[1] = helpers.Trans("terreno", player.Language.Slug)
	eTypes[2] = helpers.Trans("atmosfera", player.Language.Slug)

	//====================================
	// Validator
	//====================================
	validationFlag := false
	validationMessage := helpers.Trans("validationMessage", player.Language.Slug)
	switch state.Stage {
	case 1:
		input := message.Text
		if contains(eTypes, input) {
			payload.ExplorationType = input
			validationFlag = true
		}
	case 2:
		if time.Now().After(state.FinishAt) && payload.Times < 10 {
			log.Println("2")
			payload.Times++
			payload.Quantity = rand.Intn(3)*payload.Times + 1
			validationFlag = true
			validationMessage = helpers.Trans("wait", player.Language.Slug, state.FinishAt.Format("2006-01-02 15:04:05"))
		}
	case 3:
		input := message.Text
		if input == "Continua" /*Da tradurre*/ {
			validationFlag = true
			state.FinishAt = commands.GetEndTime(0, 10*(2*payload.Times), 0)
			state.ToNotify = true
		} else if input == "Esci" {
			state.Stage = 4
			validationFlag = true
		}
	}

	if !validationFlag {
		if state.Stage != 0 {
			validatorMsg := services.NewMessage(message.Chat.ID, validationMessage)
			services.SendMessage(validatorMsg)
		}
	}
	//====================================
	// Stage
	//====================================
	switch state.Stage {
	case 0:
		payloadUpdated, _ := json.Marshal(payloadStruct{})
		state.Payload = string(payloadUpdated)
		state.Stage = 1
		state.Update()

		msg := services.NewMessage(player.ChatID, helpers.Trans("esplorazione", player.Language.Slug))
		keyboard := make([]tgbotapi.KeyboardButton, len(eTypes))
		for i, eType := range eTypes {
			keyboard[i] = tgbotapi.NewKeyboardButton(eType)
		}
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
		services.SendMessage(msg)
	case 1:
		if validationFlag {
			// ... Recupero il materiale trovabile
			payload.Material = models.GetRandomItemByCategory(getCategoryID(message.Text, player.Language.Slug))

			// Salvo un po' tutto
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Stage = 2
			state.ToNotify = true
			// Seleziono un tipo di materiale trovabile
			state.FinishAt = commands.GetEndTime(0, 10, 0)
			msg := services.NewMessage(player.ChatID, helpers.Trans("wait", player.Language.Slug, state.FinishAt.Format("15:04:05")))
			services.SendMessage(msg)
			state.Update()
		}
	case 2:
		if validationFlag {
			payloadUpdated, _ := json.Marshal(payload)
			state.Payload = string(payloadUpdated)
			state.Stage = 3
			state.Update()
			// Hai estratto {NOME ITEM} per una quantità pari a {QUANTITA'}, vuoi terminare l'estrazione?
			msg := services.NewMessage(player.ChatID, helpers.Trans("estrazione", player.Language.Slug, payload.Material.Name, payload.Quantity))
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("continue", player.Language.Slug)), tgbotapi.NewKeyboardButton(helpers.Trans("exit", player.Language.Slug))))
			services.SendMessage(msg)
		}
	case 3:
		if validationFlag {
			state.Stage = 2
			// Setto un tempo e torno in fase 2
			msg := services.NewMessage(player.ChatID, helpers.Trans("wait", player.Language.Slug, state.FinishAt.Format("15:04:05")))
			services.SendMessage(msg)
			state.Update()
		}
	case 4:
		if validationFlag {
			helpers.FinishAndCompleteState(state, player)
			// Aggiungere item all'inventario
			player.Inventory.AddItem(payload.Material, payload.Quantity)
			player.Update()
			msg := services.NewMessage(player.ChatID, "Estrazione terminata!")
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}
	}
}

func getCategoryID(eType, lang string) uint {
	switch eType {
	case helpers.Trans("sottosuolo", lang):
		return 2
	case helpers.Trans("terreno", lang):
		return 1
	case helpers.Trans("atmosfera", lang):
		return 3
	}
	return 0
}

func contains(a []string, v string) bool {
	for _, e := range a {
		if e == v {
			return true
		}
	}
	return false
}
