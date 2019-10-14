package controllers

import (
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//====================================
// TutorialController
//====================================
type TutorialController struct {
	BaseController
	Payload struct{}
}

//====================================
// Handle
//====================================
func (c *TutorialController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	var err error
	var isNewState bool
	c.RouteName, c.Update, c.Message = "route.start", update, update.Message

	// Check current state for this routes
	c.State, isNewState = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// It's first message
	if isNewState {
		c.Stage()
		return
	}

	// Validate
	if !c.Validator() {
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}

		// Ok! Run!
		c.Stage()
		return
	}

	// Validator goes errors
	validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
	services.SendMessage(validatorMsg)
	return
}

//====================================
// Validator
//====================================
func (c *TutorialController) Validator() (hasErrors bool) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch c.State.Stage {
	case 1:
		lang, err := providers.FindLanguageBy(c.Message.Text, "name")
		if err != nil {
			services.ErrorHandler("Cant find language", err)
		}

		_, err = providers.UpdatePlayer(nnsdk.Player{ID: helpers.Player.ID, LanguageID: lang.ID})
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}

		return false
	case 2:
		if c.Message.Text == helpers.Trans("route.start.openEye") {
			return false
		}
	case 3:
		c.Validation.Message = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if c.State, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.mission"); c.State == (nnsdk.PlayerState{}) {
			return false
		}
	case 4:
		c.Validation.Message = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if c.State, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.crafting"); c.State == (nnsdk.PlayerState{}) {
			return false
		}
	case 5:
		c.Validation.Message = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if c.State, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.inventory.equip"); c.State == (nnsdk.PlayerState{}) {
			return false
		}
	case 6:
		c.Validation.Message = helpers.Trans("route.start.error.functionNotCompleted")
		// Check if the player finished the previous function.
		if c.State, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.hunting"); c.State == (nnsdk.PlayerState{}) {
			return false
		}
	}

	return true
}

//====================================
// Stage - Language -> Messages -> Exploration -> Crafting -> Hunting
//====================================
func (c *TutorialController) Stage() {
	var err error

	switch c.State.Stage {
	case 0:
		// Recupero lingue disponibili
		languages, err := providers.GetLanguages()
		if err != nil {
			services.ErrorHandler("Cant get languages", err)
		}

		// Aggiungo lingue alla tastiera
		keyboard := make([]tgbotapi.KeyboardButton, len(languages))
		for i, lang := range languages {
			keyboard[i] = tgbotapi.NewKeyboardButton(lang.Name)
		}

		// Invio messaggio
		msg := services.NewMessage(c.Message.Chat.ID, "Select language")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
		services.SendMessage(msg)

		// Aggiorna stato
		c.State.Stage = 1
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}
	case 1:
		// Messages
		texts := helpers.GenerateTextArray(c.RouteName)
		msg := services.NewMessage(helpers.Player.ChatID, texts[0])
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

		// Ultimo step apri gli occhi
		msg = services.NewMessage(helpers.Player.ChatID, "...")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("route.start.openEye"))))
		services.SendMessage(msg)

		// Aggiorna stato
		c.State.Stage = 2
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}
	case 2:
		// First Exploration
		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstExploration")))

		// Aggiorna stato
		c.State.Stage = 3
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}

		// Call mission controller
		new(MissionController).Handle(c.Update)
	case 3:
		// First Crafting
		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstCrafting")))

		// Aggiorna stato
		c.State.Stage = 4
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}

		// Call crafting controller
		new(CraftingController).Handle(c.Update)
	case 4:
		// Equip weapon
		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstWeaponEquipped")))

		// Aggiorna stato
		c.State.Stage = 5
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}

		// Call InventoryEquipController
		new(InventoryEquipController).Handle(c.Update)
	case 5:
		services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstHunting")))

		// Aggiorna stato
		c.State.Stage = 6
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}

		// Call InventoryEquipController
		new(HuntingController).Handle(c.Update)
	case 6:
		helpers.FinishAndCompleteState(c.State, helpers.Player)
	}
}
