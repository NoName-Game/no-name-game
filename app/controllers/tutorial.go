package controllers

import (
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TutorialController struct {
	RouteName  string
	Validation bool
	Update     tgbotapi.Update
	Message    *tgbotapi.Message
	Payload    struct{}
}

//====================================
// Handle
//====================================
func (c TutorialController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	c.RouteName = "route.start"
	c.Update = update
	c.Message = update.Message

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
func (c TutorialController) Validator(state nnsdk.PlayerState) (hasErrors bool, newState nnsdk.PlayerState) {
	validationMessage := helpers.Trans("validationMessage")

	switch state.Stage {
	case 1:
		lang, err := providers.FindLanguageBy(c.Message.Text, "name")
		if err != nil {
			services.ErrorHandler("Cant find language", err)
		}

		_, err = providers.UpdatePlayer(nnsdk.Player{ID: helpers.Player.ID, LanguageID: lang.ID})
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}

		return false, state
	case 2:
		if c.Message.Text == helpers.Trans("route.start.openEye") {
			return false, state
		}
	case 3:
		validationMessage = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if state, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.mission"); state == (nnsdk.PlayerState{}) {
			return false, state
		}
	case 4:
		validationMessage = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if state, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.crafting"); state == (nnsdk.PlayerState{}) {
			return false, state
		}
	case 5:
		validationMessage = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if state, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.inventory.equip"); state == (nnsdk.PlayerState{}) {
			return false, state
		}
	case 6:
		validationMessage = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if state, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.hunting"); state == (nnsdk.PlayerState{}) {
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
func (c TutorialController) Stage(state nnsdk.PlayerState) {
	//====================================
	// Language -> Messages -> Exploration -> Crafting -> Hunting
	//====================================
	switch state.Stage {
	case 0:
		msg := services.NewMessage(c.Message.Chat.ID, "Select language")

		languages, err := providers.GetLanguages()
		if err != nil {
			services.ErrorHandler("Cant get languages", err)
		}

		keyboard := make([]tgbotapi.KeyboardButton, len(languages))
		for i, lang := range languages {
			keyboard[i] = tgbotapi.NewKeyboardButton(lang.Name)
		}

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
		state.Stage = 1
		state, _ = providers.UpdatePlayerState(state)
		services.SendMessage(msg)
	case 1:
		// Messages
		texts := helpers.GenerateTextArray(c.RouteName)
		msg := services.NewMessage(helpers.Player.ChatID, texts[0])
		//msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		lastMessage := services.SendMessage(msg)
		var previousText string
		for i := 1; i < 3; i++ {
			time.Sleep(2 * time.Second)
			services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, lastMessage.MessageID, texts[i])) //.Text
		}
		for i := 3; i < 12; i++ {
			time.Sleep(2 * time.Second)
			previousText = services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, lastMessage.MessageID, previousText+"\n"+texts[i])).Text
		}
		lastMessage = services.SendMessage(services.NewMessage(helpers.Player.ChatID, texts[12]))
		previousText = lastMessage.Text
		for i := 13; i < len(texts); i++ {
			time.Sleep(time.Second)
			previousText = services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, lastMessage.MessageID, previousText+"\n"+texts[i])).Text
		}
		edit := services.NewEditMessage(helpers.Player.ChatID, lastMessage.MessageID, helpers.Trans("route.start.explosion"))
		edit.ParseMode = "HTML"
		services.SendMessage(edit)
		msg = services.NewMessage(helpers.Player.ChatID, "...")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("route.start.openEye"))))
		services.SendMessage(msg)
		state.Stage = 2
		state, _ = providers.UpdatePlayerState(state)

	case 2:
		// First Exploration
		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstExploration")))
		state.Stage = 3
		state, _ = providers.UpdatePlayerState(state)

		// Call mission controller
		new(MissionController).Handle(c.Update)
	case 3:
		// First Crafting
		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstCrafting")))
		state.Stage = 4
		state, _ = providers.UpdatePlayerState(state)
		Crafting(c.Update)
	case 4:
		// Equip weapon
		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstWeaponEquipped")))
		state.Stage = 5
		state, _ = providers.UpdatePlayerState(state)
		InventoryEquip(c.Update)
	case 5:
		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstHunting")))
		state.Stage = 6
		state, _ = providers.UpdatePlayerState(state)
		Hunting(c.Update)
	case 6:
		helpers.FinishAndCompleteState(state, helpers.Player)

	}
}
