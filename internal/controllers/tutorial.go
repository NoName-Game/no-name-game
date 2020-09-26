package controllers

import (
	"fmt"
	"time"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// TutorialController
// ====================================
type TutorialController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *TutorialController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.tutorial",
		},
	}) {
		return
	}

	// Se il player ha già finito il tutorial non può assolutamente entrare in questo controller
	if c.Player.GetTutorial() {
		c.ForceBackTo = true
		_ = c.Completing(nil)
		return
	}

	// Validate
	var hasError bool
	if hasError = c.Validator(); hasError {
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
	var err error
	switch c.CurrentState.Stage {
	case 1:
		// Verifico che l'azione passata sia quella di aprire gli occhi
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.tutorial.open_eye") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true
		}
	case 6:
		var rCheckShipTravel *pb.CheckShipTravelResponse
		if rCheckShipTravel, err = config.App.Server.Connection.CheckShipTravel(helpers.NewContext(1), &pb.CheckShipTravelRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Il crafter sta già portando a terminre un lavoro per questo player
		if !rCheckShipTravel.GetFinishTraveling() {
			var finishAt time.Time
			if finishAt, err = helpers.GetEndTime(rCheckShipTravel.GetTravelingEndTime(), c.Player); err != nil {
				c.Logger.Panic(err)
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

func (c *TutorialController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ============================================================================================================
	// Intro
	case 0:
		// Imposto start tutorial
		if _, err = config.App.Server.Connection.PlayerStartTutorial(helpers.NewContext(1), &pb.PlayerStartTutorialRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Invio set di messaggi
		// if err = c.sendIntroMessage(); err != nil {
		// 	c.Logger.Panic(err)
		// }

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
		firstUseMessage.ParseMode = "markdown"
		if _, err = helpers.SendMessage(firstUseMessage); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2

		// Richiamo inventario come sottoprocesso di questo controller
		useItemController := new(InventoryItemController)
		useItemController.Handle(c.Player, c.Update)

	// ============================================================================================================
	// Prima esplorazione
	case 2:
		// Invio messagio dove gli spiego che deve effettuare una nuova esplorazione
		firstMissionMessage := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_exploration"),
		)
		firstMissionMessage.ParseMode = "markdown"
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
		firstWeaponMessage.ParseMode = "markdown"
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
		firstHuntingMessage.ParseMode = "markdown"
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
		firstTravelMessage.ParseMode = "markdown"
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

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.travel.exploring", finishAt.Format("15:04:05 01/02")),
		)
		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 6
		c.ForceBackTo = true

	// ============================================================================================================
	// Fine viaggio
	case 6:
		// Richiamo fine viaggio
		if _, err = config.App.Server.Connection.EndShipTravel(helpers.NewContext(1), &pb.EndShipTravelRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		firstSafeMessage := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_safeplanet"),
		)
		firstSafeMessage.ParseMode = "markdown"
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
		if _, err = helpers.SendMessage(helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.completed"),
		),
		); err != nil {
			c.Logger.Panic(err)
		}

		// Registro che il player ha completato il tutorial e recupero rewward
		var rPlayerEndTutorial *pb.PlayerEndTutorialResponse
		if rPlayerEndTutorial, err = config.App.Server.Connection.PlayerEndTutorial(helpers.NewContext(1), &pb.PlayerEndTutorialRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		rewardMessage := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
			"route.tutorial.completed.reward",
			rPlayerEndTutorial.GetMoney(),
			rPlayerEndTutorial.GetExp(),
		))
		rewardMessage.ParseMode = "markdown"
		if _, err = helpers.SendMessage(rewardMessage); err != nil {
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
	initMessage := helpers.NewMessage(c.Update.Message.Chat.ID, "...")
	initMessage.ParseMode = "markdown"
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
	firstMessageConfig.ParseMode = "markdown"

	var firstMessage tgbotapi.Message
	firstMessage, err = helpers.SendMessage(firstMessageConfig)
	if err != nil {
		return err
	}

	// Mando primo set di messaggi
	for i := 1; i <= 7; i++ {
		time.Sleep(1 * time.Second)
		edited := helpers.NewEditMessage(c.Player.ChatID, firstMessage.MessageID, textList[i])
		edited.ParseMode = "markdown"

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
	secondMessageConfig.ParseMode = "markdown"

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
		edited.ParseMode = "markdown"
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
	thirdMessageConfig.ParseMode = "markdown"

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
		edited.ParseMode = "markdown"

		if thirdMessage, err = helpers.SendMessage(edited); err != nil {
			return
		}

		thirdSetText += textList[i]
	}

	// Mando messaggio allert
	time.Sleep(2 * time.Second)
	alertMessage := helpers.NewMessage(c.Update.Message.Chat.ID, textList[20])
	alertMessage.ParseMode = "markdown"
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
	fourthMessageConfig.ParseMode = "markdown"

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

		edited.ParseMode = "markdown"
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
