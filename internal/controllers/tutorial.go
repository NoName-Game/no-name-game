package controllers

import (
	"fmt"
	"time"

	"nn-grpc/build/pb"
	"nn-telegram/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-telegram/internal/helpers"
)

// ====================================
// TutorialController
// ====================================
type TutorialController struct {
	Controller
}

func (c *TutorialController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.tutorial",
		},
		Configurations: ControllerConfigurations{
			AllowedControllers: []string{
				"route.exploration",
				"route.ship.travel.finding",
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *TutorialController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Se il player ha già finito il tutorial non può assolutamente entrare in questo controller
	if c.Player.GetTutorial() {
		c.ForceBackTo = true
		c.Completing(nil)
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
	c.Completing(nil)
}

// ====================================
// Validator
// ====================================
func (c *TutorialController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico che l'azione passata sia quella di aprire gli occhi
	// ##################################################################################################
	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.tutorial.open_eye") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true
		}
	// ##################################################################################################
	// Verifico se il player ha compleato le attività
	// ##################################################################################################
	case 3:
		for _, state := range c.Data.PlayerActiveStates {
			if state.Controller != "route.tutorial" {
				c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")
				return true
			}
		}
	// ##################################################################################################
	// Verifico se il player ha compleato le attività
	// ##################################################################################################
	case 6:
		for _, state := range c.Data.PlayerActiveStates {
			if state.Controller != "route.tutorial" {
				c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")
				return true
			}
		}
	}

	return false
}

func (c *TutorialController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ============================================================================================================
	// Intro
	case 0:
		// Invio set di messaggi
		if err = c.sendIntroMessage(); err != nil {
			c.Logger.Panic(err)
		}

		// Ultimo step apri gli occhi
		var openEyeMessage tgbotapi.MessageConfig
		openEyeMessage = helpers.NewMessage(c.Player.ChatID, "...")
		openEyeMessage.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.tutorial.open_eye"),
				),
			),
		)
		if _, err = helpers.SendMessage(openEyeMessage); err != nil {
			return
		}

		// Aggiorna stato
		c.CurrentState.Stage = 1

	// ============================================================================================================
	// Primo uso item
	case 1:
		// Invio messagio dove gli spiego come usare gli item
		firstUseMessage := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_use_item"),
		)
		firstUseMessage.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(firstUseMessage); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2

		// Richiamo inventario come sottoprocesso di questo controller
		useItemController := new(PlayerInventoryController)
		useItemController.Handle(c.Player, c.Update)

	// ============================================================================================================
	// Prima esplorazione
	case 2:
		// Invio messagio dove gli spiego che deve effettuare una nuova esplorazione
		firstMissionMessage := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_exploration"),
		)
		firstMissionMessage.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(firstMissionMessage); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3

		// Richiamo esplorazione come sottoprocesso di questo controller
		missionController := new(ExplorationController)
		missionController.Handle(c.Player, c.Update)

	// ============================================================================================================
	// Equipaggiamento arma
	case 3:
		firstWeaponMessage := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_weapon_equipped"),
		)
		firstWeaponMessage.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(firstWeaponMessage); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 4

		// Richiamo equipment come sottoprocesso di questo controller
		inventoryController := new(PlayerEquipmentController)
		inventoryController.Handle(c.Player, c.Update)

	// ============================================================================================================
	// Hunting
	case 4:
		firstHuntingMessage := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_hunting"),
		)
		firstHuntingMessage.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(firstHuntingMessage); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 5

		// Richiamo huting come sottoprocesso di questo controller
		huntingController := new(HuntingController)
		huntingController.Handle(c.Player, c.Update)

	// ============================================================================================================
	// Primo viaggio verso pianeta sicuro
	case 5:
		// Questo stage fa viaggiare il player forzatamente verso un pianeta sicuro
		firstTravelMessage := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_travel"),
		)
		firstTravelMessage.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(firstTravelMessage); err != nil {
			c.Logger.Panic(err)
		}

		// Player start travel
		var rStartTravelTutorial *pb.StartTravelTutorialResponse
		if rStartTravelTutorial, err = config.App.Server.Connection.StartTravelTutorial(helpers.NewContext(1), &pb.StartTravelTutorialRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero orario fine viaggio
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rStartTravelTutorial.GetTravelingEndTime(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		// Creo cache per lo stato di viaggio
		if err = helpers.SetControllerCacheData(c.Player.ID, "route.ship.travel.finding", 2, nil); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID,
			helpers.Trans(c.Player.Language.Slug, "ship.travel.exploring", finishAt.Format("15:04:05 01/02")),
		)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 6
		c.ForceBackTo = true

	// ============================================================================================================
	// Fine viaggio
	case 6:
		firstSafeMessage := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_safeplanet"),
		)
		firstSafeMessage.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(firstSafeMessage); err != nil {
			c.Logger.Panic(err)
		}

		// Forzo a mano l'aggiornamento dello stato del player
		// in quanto adesso devo richiamare un'altro controller
		c.CurrentState.Stage = 7
		c.ForceBackTo = true

	// ============================================================================================================
	// Tutorial completato
	case 7:
		// Registro che il player ha completato il tutorial e recupero rewward
		if _, err = config.App.Server.Connection.PlayerEndTutorial(helpers.NewContext(1), &pb.PlayerEndTutorialRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		if _, err = helpers.SendMessage(helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.completed"),
		),
		); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}

func (c *TutorialController) sendIntroMessage() (err error) {
	// Recupero set di messaggi
	textList := helpers.GenerateTextArray(c.Player.Language.Slug, c.CurrentState.Controller)

	// Invio il primo messaggio per eliminare la tastiera
	initMessage := helpers.NewMessage(c.ChatID, "...")
	initMessage.ParseMode = tgbotapi.ModeHTML
	initMessage.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err = helpers.SendMessage(initMessage); err != nil {
		return err
	}

	// ************************
	// Primo set di messaggi
	// ************************
	time.Sleep(1 * time.Second)

	var firstMessageConfig tgbotapi.MessageConfig
	firstMessageConfig = helpers.NewMessage(c.Player.ChatID, textList[0])
	firstMessageConfig.ParseMode = tgbotapi.ModeHTML

	var firstMessage tgbotapi.Message
	firstMessage, err = helpers.SendMessage(firstMessageConfig)
	if err != nil {
		return err
	}

	// Mando primo set di messaggi
	for i := 1; i <= 7; i++ {
		time.Sleep(1 * time.Second)
		edited := helpers.NewEditMessage(c.Player.ChatID, firstMessage.MessageID, textList[i])
		edited.ParseMode = tgbotapi.ModeHTML

		if _, err = helpers.SendMessage(edited); err != nil {
			return err
		}
	}

	// ************************
	// Secondo set di messaggi
	// ************************
	time.Sleep(1 * time.Second)
	var secondSetText = textList[8]

	var secondMessageConfig tgbotapi.MessageConfig
	secondMessageConfig = helpers.NewMessage(c.Player.ChatID, secondSetText)
	secondMessageConfig.ParseMode = tgbotapi.ModeHTML

	var secondMessage tgbotapi.Message
	if secondMessage, err = helpers.SendMessage(secondMessageConfig); err != nil {
		return err
	}

	// PreviusText mi serve per andare a modificare il messaggio
	// inviato ed appendergli la nuova parte di messaggio
	for i := 9; i <= 12; i++ {
		time.Sleep(2 * time.Second)
		currentMessage := fmt.Sprintf("%s%s", secondSetText, textList[i])

		edited := helpers.NewEditMessage(
			c.Player.ChatID,
			secondMessage.MessageID,
			currentMessage,
		)
		edited.ParseMode = tgbotapi.ModeHTML
		if secondMessage, err = helpers.SendMessage(edited); err != nil {
			return
		}

		// Concateno messaggi
		secondSetText += textList[i]
	}

	// ************************
	// Terzo set di messaggi
	// ************************
	time.Sleep(1 * time.Second)
	thirdSetText := textList[13]

	var thirdMessageConfig tgbotapi.MessageConfig
	thirdMessageConfig = helpers.NewMessage(c.Player.ChatID, thirdSetText)
	thirdMessageConfig.ParseMode = tgbotapi.ModeHTML

	var thirdMessage tgbotapi.Message
	if thirdMessage, err = helpers.SendMessage(thirdMessageConfig); err != nil {
		return err
	}

	// PreviusText mi serve per andare a modificare il messaggio
	// inviato ed appendergli la nuova parte di messaggio
	for i := 14; i <= 19; i++ {
		currentMessage := fmt.Sprintf("%s%s", thirdSetText, textList[i])

		time.Sleep(2 * time.Second)
		edited := helpers.NewEditMessage(
			c.Player.ChatID,
			thirdMessage.MessageID,
			currentMessage,
		)
		edited.ParseMode = tgbotapi.ModeHTML

		if thirdMessage, err = helpers.SendMessage(edited); err != nil {
			return
		}

		thirdSetText += textList[i]
	}

	// Mando messaggio allert
	time.Sleep(2 * time.Second)
	alertMessage := helpers.NewMessage(c.ChatID, textList[20])
	alertMessage.ParseMode = tgbotapi.ModeHTML
	alertMessage.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err = helpers.SendMessage(alertMessage); err != nil {
		return err
	}

	// ************************
	// Quarto set di messaggi ( COUNTDOWN )
	// ************************
	time.Sleep(2 * time.Second)
	var fourthMessageConfig tgbotapi.MessageConfig
	fourthMessageConfig = helpers.NewMessage(c.Player.ChatID, textList[21])
	fourthMessageConfig.ParseMode = tgbotapi.ModeHTML

	var fourthMessage tgbotapi.Message
	if fourthMessage, err = helpers.SendMessage(fourthMessageConfig); err != nil {
		return err
	}

	// Mando primo set di messaggi
	for i := 22; i <= 27; i++ {
		time.Sleep(1 * time.Second)
		edited := helpers.NewEditMessage(
			c.Player.ChatID,
			fourthMessage.MessageID,
			textList[i],
		)

		edited.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(edited); err != nil {
			return err
		}
	}

	// ************************
	// Esplosione
	// ************************
	edit := helpers.NewEditMessage(
		c.Player.ChatID,
		fourthMessage.MessageID,
		helpers.Trans(c.Player.Language.Slug, "route.tutorial.explosion"),
	)

	edit.ParseMode = "HTML"
	if _, err = helpers.SendMessage(edit); err != nil {
		return
	}

	return
}
