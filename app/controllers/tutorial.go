package controllers

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// TutorialController
// ====================================
type TutorialController struct {
	BaseController
	Payload struct{}
}

// ====================================
// Handle
// ====================================
func (c *TutorialController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Se il player ha già finito il tutorial non può assolutamente entrare in questo controller
	if c.Player.GetTutorial() {
		c.ForceBackTo = true
		_ = c.Completing(c.Payload)
		return
	}

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller: "route.tutorial",
		Payload:    c.Payload,
	}) {
		return
	}

	// Carico payload
	if err = helpers.GetPayloadController(c.Player.ID, c.CurrentState.Controller, &c.Payload); err != nil {
		panic(err)
	}

	// Validate
	var hasError bool
	if hasError = c.Validator(); hasError {
		c.Validate()
		return
	}

	// Ok! Run!
	if err = c.Stage(); err != nil {
		panic(err)
	}

	// Completo progressione
	if err = c.Completing(&c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *TutorialController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	case 1:
		// Recupero lingue disponibili
		if _, err = services.NnSDK.GetLanguageByName(helpers.NewContext(1), &pb.GetLanguageByNameRequest{
			Name: c.Update.Message.Text,
		}); err != nil {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true
		}
	case 2:
		// Verifico che l'azione passata sia quella di aprire gli occhi
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.tutorial.open_eye") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true
		}
	case 7:
		var rCheckShipTravel *pb.CheckShipTravelResponse
		if rCheckShipTravel, err = services.NnSDK.CheckShipTravel(helpers.NewContext(1), &pb.CheckShipTravelRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			panic(err)
		}

		// Il crafter sta già portando a terminre un lavoro per questo player
		if !rCheckShipTravel.GetFinishTraveling() {
			var finishAt time.Time
			if finishAt, err = ptypes.Timestamp(rCheckShipTravel.GetTravelingEndTime()); err != nil {
				panic(err)
			}

			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"ship.travel.wait",
				finishAt.Format("15:04:05"),
			)

			return true
		}
	}

	return false
}

func (c *TutorialController) Stage() (err error) {
	switch c.CurrentState.Stage {
	// ============================================================================================================
	// Settaggio lingua
	case 0:
		// Recupero lingue disponibili
		var rGetLanguages *pb.GetAllLanguagesResponse
		rGetLanguages, err = services.NnSDK.GetAllLanguages(helpers.NewContext(1), &pb.GetAllLanguagesRequest{})
		if err != nil {
			return err
		}

		// Aggiungo lingue alla tastiera
		keyboard := make([]tgbotapi.KeyboardButton, len(rGetLanguages.GetLanguages()))
		for i, lang := range rGetLanguages.GetLanguages() {
			keyboard[i] = tgbotapi.NewKeyboardButton(lang.Name)
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "select_language"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorna stato
		c.CurrentState.Stage = 1

	// ============================================================================================================
	// Intro
	case 1:
		// Imposto start tutorial
		if _, err = services.NnSDK.PlayerStartTutorial(helpers.NewContext(1), &pb.PlayerStartTutorialRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			return err
		}

		// Invio set di messaggi
		// if err = c.sendIntroMessage(); err != nil {
		// 	return err
		// }

		// Ultimo step apri gli occhi
		var openEyeMessage tgbotapi.MessageConfig
		openEyeMessage = services.NewMessage(c.Player.ChatID, "...")
		openEyeMessage.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.tutorial.open_eye"),
				),
			),
		)
		if _, err = services.SendMessage(openEyeMessage); err != nil {
			return
		}

		// Aggiorna stato
		c.CurrentState.Stage = 2

	// ============================================================================================================
	// Primo uso item
	case 2:
		// Invio messagio dove gli spiego come usare gli item
		firstUseMessage := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_use_item"),
		)
		firstUseMessage.ParseMode = "markdown"
		if _, err = services.SendMessage(firstUseMessage); err != nil {
			return
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3

		// Richiamo inventario come sottoprocesso di questo controller
		useItemController := new(InventoryItemController)
		useItemController.Handle(c.Player, c.Update)

	// ============================================================================================================
	// Prima esplorazione
	case 3:
		// Invio messagio dove gli spiego che deve effettuare una nuova esplorazione
		firstMissionMessage := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_exploration"),
		)
		firstMissionMessage.ParseMode = "markdown"
		if _, err = services.SendMessage(firstMissionMessage); err != nil {
			return
		}

		// Aggiorno stato
		c.CurrentState.Stage = 4

		// Richiamo esplorazione come sottoprocesso di questo controller
		missionController := new(ExplorationController)
		missionController.Handle(c.Player, c.Update)

	// ============================================================================================================
	// Equipaggiamento arma
	case 4:
		firstWeaponMessage := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_weapon_equipped"),
		)
		firstWeaponMessage.ParseMode = "markdown"
		if _, err = services.SendMessage(firstWeaponMessage); err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 5

		// Richiamo equipment come sottoprocesso di questo controller
		inventoryController := new(PlayerEquipmentController)
		inventoryController.Handle(c.Player, c.Update)

	// ============================================================================================================
	// Hunting
	case 5:
		firstHuntingMessage := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_hunting"),
		)
		firstHuntingMessage.ParseMode = "markdown"
		if _, err = services.SendMessage(firstHuntingMessage); err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 6

		// Richiamo huting come sottoprocesso di questo controller
		huntingController := new(HuntingController)
		huntingController.Handle(c.Player, c.Update)

	// ============================================================================================================
	// Primo viaggio verso pianeta sicuro
	case 6:
		// Questo stage fa viaggiare il player forzatamente verso un pianeta sicuro
		firstTravelMessage := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_travel"),
		)
		firstTravelMessage.ParseMode = "markdown"
		if _, err = services.SendMessage(firstTravelMessage); err != nil {
			return err
		}

		// Player start travel
		var rStartTravelTutorial *pb.StartTravelTutorialResponse
		if rStartTravelTutorial, err = services.NnSDK.StartTravelTutorial(helpers.NewContext(1), &pb.StartTravelTutorialRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			return err
		}

		// Recupero orario fine viaggio
		var finishAt time.Time
		if finishAt, err = ptypes.Timestamp(rStartTravelTutorial.GetTravelingEndTime()); err != nil {
			return
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.travel.exploring", finishAt.Format("15:04:05 01/02")),
		)
		msg.ParseMode = "markdown"
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 7
		c.ForceBackTo = true

	// ============================================================================================================
	// Fine viaggio
	case 7:
		// Richiamo fine viaggio
		if _, err = services.NnSDK.EndShipTravel(helpers.NewContext(1), &pb.EndShipTravelRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			return
		}

		firstSafeMessage := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_safeplanet"),
		)
		firstSafeMessage.ParseMode = "markdown"
		if _, err = services.SendMessage(firstSafeMessage); err != nil {
			return err
		}

		// Forzo a mano l'aggiornamento dello stato del player
		// in quanto adesso devo richiamare un'altro controller
		c.CurrentState.Stage = 8
		c.ForceBackTo = true

	// ============================================================================================================
	// Tutorial completato
	case 8:
		if _, err = services.SendMessage(services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.completed"),
		),
		); err != nil {
			return err
		}

		// Registro che il player ha completato il tutorial e recupero rewward
		var rPlayerEndTutorial *pb.PlayerEndTutorialResponse
		if rPlayerEndTutorial, err = services.NnSDK.PlayerEndTutorial(helpers.NewContext(1), &pb.PlayerEndTutorialRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			return
		}

		rewardMessage := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
			"route.tutorial.completed.reward",
			rPlayerEndTutorial.GetMoney(),
			rPlayerEndTutorial.GetExp(),
		))
		rewardMessage.ParseMode = "markdown"
		if _, err = services.SendMessage(rewardMessage); err != nil {
			return err
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}

func (c *TutorialController) sendIntroMessage() (err error) {
	// Recupero set di messaggi
	textList := helpers.GenerateTextArray(c.Player.Language.Slug, c.Configuration.Controller)

	// Invio il primo messaggio per eliminare la tastiera
	initMessage := services.NewMessage(c.Update.Message.Chat.ID, "...")
	initMessage.ParseMode = "markdown"
	initMessage.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err = services.SendMessage(initMessage); err != nil {
		return err
	}

	// ************************
	// Primo set di messaggi
	// ************************
	time.Sleep(1 * time.Second)

	var firstMessageConfig tgbotapi.MessageConfig
	firstMessageConfig = services.NewMessage(c.Player.ChatID, textList[0])
	firstMessageConfig.ParseMode = "markdown"

	var firstMessage tgbotapi.Message
	firstMessage, err = services.SendMessage(firstMessageConfig)
	if err != nil {
		return err
	}

	// Mando primo set di messaggi
	for i := 1; i <= 7; i++ {
		time.Sleep(1 * time.Second)
		edited := services.NewEditMessage(c.Player.ChatID, firstMessage.MessageID, textList[i])
		edited.ParseMode = "markdown"

		if _, err = services.SendMessage(edited); err != nil {
			return err
		}
	}

	// ************************
	// Secondo set di messaggi
	// ************************
	time.Sleep(1 * time.Second)
	var secondSetText = textList[8]

	var secondMessageConfig tgbotapi.MessageConfig
	secondMessageConfig = services.NewMessage(c.Player.ChatID, secondSetText)
	secondMessageConfig.ParseMode = "markdown"

	var secondMessage tgbotapi.Message
	if secondMessage, err = services.SendMessage(secondMessageConfig); err != nil {
		return err
	}

	// PreviusText mi serve per andare a modificare il messaggio
	// inviato ed appendergli la nuova parte di messaggio
	for i := 9; i <= 12; i++ {
		time.Sleep(2 * time.Second)
		currentMessage := fmt.Sprintf("%s%s", secondSetText, textList[i])

		edited := services.NewEditMessage(
			c.Player.ChatID,
			secondMessage.MessageID,
			currentMessage,
		)
		edited.ParseMode = "markdown"
		if secondMessage, err = services.SendMessage(edited); err != nil {
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
	thirdMessageConfig = services.NewMessage(c.Player.ChatID, thirdSetText)
	thirdMessageConfig.ParseMode = "markdown"

	var thirdMessage tgbotapi.Message
	if thirdMessage, err = services.SendMessage(thirdMessageConfig); err != nil {
		return err
	}

	// PreviusText mi serve per andare a modificare il messaggio
	// inviato ed appendergli la nuova parte di messaggio
	for i := 14; i <= 19; i++ {
		currentMessage := fmt.Sprintf("%s%s", thirdSetText, textList[i])

		time.Sleep(2 * time.Second)
		edited := services.NewEditMessage(
			c.Player.ChatID,
			thirdMessage.MessageID,
			currentMessage,
		)
		edited.ParseMode = "markdown"

		if thirdMessage, err = services.SendMessage(edited); err != nil {
			return
		}

		thirdSetText += textList[i]
	}

	// Mando messaggio allert
	time.Sleep(2 * time.Second)
	alertMessage := services.NewMessage(c.Update.Message.Chat.ID, textList[20])
	alertMessage.ParseMode = "markdown"
	alertMessage.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err = services.SendMessage(alertMessage); err != nil {
		return err
	}

	// ************************
	// Quarto set di messaggi ( COUNTDOWN )
	// ************************
	time.Sleep(2 * time.Second)
	var fourthMessageConfig tgbotapi.MessageConfig
	fourthMessageConfig = services.NewMessage(c.Player.ChatID, textList[21])
	fourthMessageConfig.ParseMode = "markdown"

	var fourthMessage tgbotapi.Message
	if fourthMessage, err = services.SendMessage(fourthMessageConfig); err != nil {
		return err
	}

	// Mando primo set di messaggi
	for i := 22; i <= 27; i++ {
		time.Sleep(1 * time.Second)
		edited := services.NewEditMessage(
			c.Player.ChatID,
			fourthMessage.MessageID,
			textList[i],
		)

		edited.ParseMode = "markdown"
		if _, err = services.SendMessage(edited); err != nil {
			return err
		}
	}

	// ************************
	// Esplosione
	// ************************
	edit := services.NewEditMessage(
		c.Player.ChatID,
		fourthMessage.MessageID,
		helpers.Trans(c.Player.Language.Slug, "route.tutorial.explosion"),
	)

	edit.ParseMode = "HTML"
	if _, err = services.SendMessage(edit); err != nil {
		return
	}

	return
}
