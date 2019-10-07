package controllers

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type MissionController struct {
	RouteName  string
	Validation bool
	Update     tgbotapi.Update
	Message    *tgbotapi.Message
	Payload    struct {
		ExplorationType string
		Times           int
		Material        nnsdk.Resource
		Quantity        int
	}
	// Additional Data
	MissionTypes []string
}

//====================================
// Handle
//====================================
func (c MissionController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	c.RouteName = "route.mission"
	c.Update = update
	c.Message = update.Message

	// Set mission types
	c.MissionTypes = make([]string, 3)
	c.MissionTypes[0] = helpers.Trans("mission.underground")
	c.MissionTypes[1] = helpers.Trans("mission.surface")
	c.MissionTypes[2] = helpers.Trans("mission.atmosphere")

	// Check current state for this routes
	state, isNewState := helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// It's first message
	if isNewState {
		c.Stage(state)
		return
	}

	// Go to validator
	c.Validation, state = c.Validator(state)
	if !c.Validation {
		state, _ = providers.UpdatePlayerState(state)
		c.Stage(state)
	}

	return
}

//====================================
// Validator
//====================================
func (c MissionController) Validator(state nnsdk.PlayerState) (hasErrors bool, newState nnsdk.PlayerState) {
	validationMessage := helpers.Trans("validationMessage")

	switch state.Stage {
	case 1:
		if helpers.StringInSlice(c.Message.Text, c.MissionTypes) {
			c.Payload.ExplorationType = c.Message.Text
			return false, state
		}
	case 2:
		validationMessage = helpers.Trans("mission.wait", state.FinishAt.Format("2006-01-02 15:04:05"))
		if time.Now().After(state.FinishAt) && c.Payload.Times < 10 {
			c.Payload.Times++
			c.Payload.Quantity = rand.Intn(3)*c.Payload.Times + 1
			return false, state
		}
	case 3:
		if c.Message.Text == helpers.Trans("mission.continue") {
			state.FinishAt = helpers.GetEndTime(0, 10*(2*c.Payload.Times), 0)
			t := new(bool)
			*t = true
			state.ToNotify = t
			return false, state
		} else if c.Message.Text == helpers.Trans("mission.comeback") {
			state.Stage = 4
			return false, state
		}
	}

	// Validator goes errors
	validatorMsg := services.NewMessage(c.Message.Chat.ID, validationMessage)
	services.SendMessage(validatorMsg)

	return true, state
}

//====================================
// Stage
//====================================
func (c MissionController) Stage(state nnsdk.PlayerState) {
	switch state.Stage {
	case 0:
		state.Stage = 1
		state, _ = providers.UpdatePlayerState(state)

		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("mission.exploration"))

		var keyboardRows [][]tgbotapi.KeyboardButton
		for _, eType := range c.MissionTypes {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(eType))
			keyboardRows = append(keyboardRows, keyboardRow)
		}
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRows,
			ResizeKeyboard: true,
		}
		services.SendMessage(msg)
	case 1:
		material, err := providers.GetRandomResource(helpers.GetMissionCategoryID(c.Message.Text))
		if err != nil {
			log.Println(err)
			services.ErrorHandler("Cant add resource to player inventory", err)
		}

		// Updating state
		c.Payload.Material = material
		jsonPayload, _ := json.Marshal(c.Payload)
		state.Payload = string(jsonPayload)
		state.Stage = 2

		t := new(bool)
		*t = true
		state.ToNotify = t

		// Set finishAt
		state.FinishAt = helpers.GetEndTime(0, 10, 0)
		state, _ = providers.UpdatePlayerState(state)

		// Remove current redist stare
		helpers.DelRedisState(helpers.Player)

		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("mission.wait", string(state.FinishAt.Format("15:04:05"))))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("route.Menu"))))
		services.SendMessage(msg)

	case 2:
		jsonPayload, _ := json.Marshal(c.Payload)
		state.Payload = string(jsonPayload)
		state.Stage = 3
		state, _ = providers.UpdatePlayerState(state)

		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("mission.extraction_recap", c.Payload.Material.Name, c.Payload.Quantity))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("mission.continue")),
				tgbotapi.NewKeyboardButton(helpers.Trans("mission.comeback")),
			),
		)
		services.SendMessage(msg)

	case 3:
		state.Stage = 2
		state, _ = providers.UpdatePlayerState(state)

		// Remove current redist stare
		helpers.DelRedisState(helpers.Player)

		// Continue
		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("mission.wait", string(state.FinishAt.Format("15:04:05"))))
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		services.SendMessage(msg)
	case 4:
		// Exit
		helpers.FinishAndCompleteState(state, helpers.Player)

		_, err := providers.AddResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
			ItemID:   c.Payload.Material.ID,
			Quantity: c.Payload.Quantity,
		})

		if err != nil {
			services.ErrorHandler("Cant add resource to player inventory", err)
		}

		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("mission.extraction_ended"))
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		services.SendMessage(msg)
	}
}
