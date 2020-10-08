package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SetupController
// ====================================
type SetupController struct {
	Controller
	Paylaod struct {
		LanguageID uint32
		TimezoneID uint32
	}
}

// ====================================
// Handle
// ====================================
func (c *SetupController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.setup",
			Payload:    &c.Paylaod,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To: &MenuController{},
			},
		},
	}) {
		return
	}

	// Validate
	if c.Validator() {
		c.Validate()
		return
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(&c.Paylaod)
}

// ====================================
// Validator
// ====================================
func (c *SetupController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifica timezone
	// ##################################################################################################
	case 1:
		var rGetTimezoneByDescription *pb.GetTimezoneByDescriptionResponse
		if rGetTimezoneByDescription, err = config.App.Server.Connection.GetTimezoneByDescription(helpers.NewContext(1), &pb.GetTimezoneByDescriptionRequest{
			Description: c.Update.Message.Text,
		}); err != nil {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true
		}

		// Ho trovato il timezone
		c.Paylaod.TimezoneID = rGetTimezoneByDescription.GetTimezone().GetID()
	// ##################################################################################################
	// Verifico se la lingua scelta esiste
	// ##################################################################################################
	case 2:
		var rGetLanguageByName *pb.GetLanguageByNameResponse
		if rGetLanguageByName, err = config.App.Server.Connection.GetLanguageByName(helpers.NewContext(1), &pb.GetLanguageByNameRequest{
			Name: c.Update.Message.Text,
		}); err != nil {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true
		}

		// Ho trovato la lingua
		c.Paylaod.LanguageID = rGetLanguageByName.GetLanguage().GetID()

	}

	return false
}

func (c *SetupController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ============================================================================================================
	// Settaggio lingua
	case 0:
		// Invio messaggio
		msgIntro := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "setup.intro", c.Player.Username))
		msgIntro.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msgIntro); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero lista timezones
		var rGetAllTimezones *pb.GetAllTimezonesResponse
		if rGetAllTimezones, err = config.App.Server.Connection.GetAllTimezones(helpers.NewContext(1), &pb.GetAllTimezonesRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		// Keyboard con riassunto risorse necessarie
		var keyboard [][]tgbotapi.KeyboardButton
		for _, timezone := range rGetAllTimezones.GetTimezones() {
			keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
				timezone.GetDescription(),
			)))
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "setup.select_timezone"))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorna stato
		c.CurrentState.Stage = 1

	// ============================================================================================================
	// Salvo lingua e chiedo timezone
	case 1:
		// Aggiorno timezone scelto dal player
		if _, err = config.App.Server.Connection.PlayerSetTimezone(helpers.NewContext(1), &pb.PlayerSetTimezoneRequest{
			PlayerID:   c.Player.ID,
			TimezoneID: c.Paylaod.TimezoneID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero lingue disponibili
		var rGetLanguages *pb.GetAllLanguagesResponse
		if rGetLanguages, err = config.App.Server.Connection.GetAllLanguages(helpers.NewContext(1), &pb.GetAllLanguagesRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiungo lingue alla tastiera
		keyboard := make([]tgbotapi.KeyboardButton, len(rGetLanguages.GetLanguages()))
		for i, lang := range rGetLanguages.GetLanguages() {
			keyboard[i] = tgbotapi.NewKeyboardButton(lang.Name)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "setup.select_language"))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorna stato
		c.CurrentState.Stage = 2

	// ============================================================================================================
	// Completo configurazione
	case 2:
		// Aggiorno lingua scelta dal player
		if _, err = config.App.Server.Connection.PlayerSetLanguage(helpers.NewContext(1), &pb.PlayerSetLanguageRequest{
			PlayerID:   c.Player.ID,
			LanguageID: c.Paylaod.LanguageID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Mando messaggio fine confiurazione
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "setup.end"))
		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Se il player non ha mai eseguito il tutorial lo mando li
		if !c.Player.Tutorial {
			c.Configurations.ControllerBack.To = &TutorialController{}
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
