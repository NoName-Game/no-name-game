package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Tutorial:
// Tutorial iniziale fake per introdurre il player alle meccaniche base di NoName.
// Flow: Atterraggio d'emergenza -> ricerca materiali per riparare nave -> semplice crafting ->
// hunting (?) -> volo nel sistema di spawn -> Fine Tutorial

// ====================================
// TutorialController
// ====================================
type TutorialController struct {
	BaseController
	Payload struct {
		UseItemID        uint
		MissionID        uint
		CraftingID       uint
		HuntingID        uint
		InventoryEquipID uint
	}
}

// ====================================
// Handle
// ====================================
func (c *TutorialController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Controller = "route.tutorial"
	c.Player = player
	c.Update = update

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	// Stato recuperto correttamente
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// Validate
	var hasError bool
	hasError, err = c.Validator()
	if err != nil {
		panic(err)
	}

	// Se ritornano degli errori
	if hasError == true {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
		validatorMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
				),
			),
		)

		_, err = services.SendMessage(validatorMsg)
		if err != nil {
			panic(err)
		}

		return
	}

	// Ok! Run!
	err = c.Stage()
	if err != nil {
		panic(err)
	}

	// Aggiorno stato finale
	payloadUpdated, _ := json.Marshal(c.Payload)
	c.State.Payload = string(payloadUpdated)
	_, err = providers.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}

	return
}

// ====================================
// Validator
// ====================================
func (c *TutorialController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	switch c.State.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, nil

	// In questo stage è necessario controllare se la lingua passata è quella giusta
	case 1:
		// Recupero lingue disponibili
		_, err := providers.FindLanguageBy(c.Update.Message.Text, "name")

		// Verifico se la lingua esiste, se così non fosse ritorno errore
		if err != nil {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true, nil
		}

		return false, nil

	// In questo stage devo verificare unicamente che venga passata una stringa
	case 2:
		// Verifico che l'azione passata sia quella di aprire gli occhi
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.tutorial.open_eye") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true, nil
		}

		return false, nil

	// In questo stage verifico se il player ha usato il rivitalizzante
	case 3:
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")

		var stateNotFoundErr error
		var missionState nnsdk.PlayerState
		missionState, stateNotFoundErr = providers.GetPlayerStateByID(c.Payload.UseItemID)
		// Non è stato trovato lo stato ritorno allo stato precedente
		// e non ritorno errore
		if stateNotFoundErr != nil {
			c.State.Stage = 2
			return false, err
		}

		if *missionState.Completed != true {
			return true, err
		}

		return false, err

	// In questo stage verifico se il player ha completato correttamente la missione
	case 5:
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")

		var stateNotFoundErr error
		var missionState nnsdk.PlayerState
		missionState, stateNotFoundErr = providers.GetPlayerStateByID(c.Payload.MissionID)
		// Non è stato trovato lo stato ritorno allo stato precedente
		// e non ritorno errore
		if stateNotFoundErr != nil {
			c.State.Stage = 3
			return false, err
		}

		if *missionState.Completed != true {
			return true, err
		}

		return false, err
	// case 5:
	// 	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")
	//
	// 	var stateNotFoundErr error
	// 	var missionState nnsdk.PlayerState
	// 	missionState, stateNotFoundErr = providers.GetPlayerStateByID(c.Payload.CraftingID)
	// 	// Non è stato trovato lo stato ritorno allo stato precedente
	// 	// e non ritorno errore
	// 	if stateNotFoundErr != nil {
	// 		c.State.Stage = 4
	// 		return false, err
	// 	}
	//
	// 	if *missionState.Completed != true {
	// 		return true, err
	// 	}
	//
	// 	return false, err
	case 6:
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")

		var stateNotFoundErr error
		var missionState nnsdk.PlayerState
		missionState, stateNotFoundErr = providers.GetPlayerStateByID(c.Payload.InventoryEquipID)
		// Non è stato trovato lo stato ritorno allo stato precedente
		// e non ritorno errore
		if stateNotFoundErr != nil {
			c.State.Stage = 5
			return false, err
		}

		if *missionState.Completed != true {
			return true, err
		}

		return false, err
	case 7:
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")

		var stateNotFoundErr error
		var missionState nnsdk.PlayerState
		missionState, stateNotFoundErr = providers.GetPlayerStateByID(c.Payload.HuntingID)
		// Non è stato trovato lo stato ritorno allo stato precedente
		// e non ritorno errore
		if stateNotFoundErr != nil {
			c.State.Stage = 6
			return false, err
		}

		if *missionState.Completed != true {
			return true, err
		}

		return false, err
	default:
		// Stato non riconosciuto ritorno errore
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.state")
	}

	return true, err
}

// ====================================
// Stage - Language -> Messages -> Exploration -> Crafting -> Hunting
// ====================================
func (c *TutorialController) Stage() (err error) {
	switch c.State.Stage {
	// Primo avvio in questo momento l'utente deve poter ricevere la lista delle lingue disponibili
	// e potrà selezionare la sua lingua tramite tastierino
	case 0:
		// Recupero lingue disponibili
		languages, err := providers.GetLanguages()
		if err != nil {
			return err
		}

		// Aggiungo lingue alla tastiera
		keyboard := make([]tgbotapi.KeyboardButton, len(languages))
		for i, lang := range languages {
			keyboard[i] = tgbotapi.NewKeyboardButton(lang.Name)
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, "Select language")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorna stato
		c.State.Stage = 1

	// In questo stage è previsto un'invio di un set di messaggi
	// che introducono al player cosa sta accadendo
	case 1:
		// Recupero set di messaggi
		textList := helpers.GenerateTextArray(c.Player.Language.Slug, c.Controller)

		// Invio il primo messaggio per eliminare la tastiera
		initMessage := services.NewMessage(c.Update.Message.Chat.ID, "...")
		initMessage.ParseMode = "markdown"
		initMessage.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, err = services.SendMessage(initMessage)
		if err != nil {
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

			_, err := services.SendMessage(edited)
			if err != nil {
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
		secondMessage, err = services.SendMessage(secondMessageConfig)
		if err != nil {
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

			secondMessage, err = services.SendMessage(edited)
			if err != nil {
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
		thirdMessage, err = services.SendMessage(thirdMessageConfig)
		if err != nil {
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

			thirdMessage, err = services.SendMessage(edited)
			if err != nil {
				return
			}

			thirdSetText += textList[i]
		}

		// Mando messaggio allert
		time.Sleep(1 * time.Second)
		alertMessage := services.NewMessage(c.Update.Message.Chat.ID, textList[20])
		alertMessage.ParseMode = "markdown"
		alertMessage.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		_, err = services.SendMessage(alertMessage)
		if err != nil {
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
		fourthMessage, err = services.SendMessage(fourthMessageConfig)
		if err != nil {
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
			_, err := services.SendMessage(edited)
			if err != nil {
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
		_, err = services.SendMessage(edit)
		if err != nil {
			return
		}

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
		_, err = services.SendMessage(openEyeMessage)
		if err != nil {
			return
		}

		// Aggiorna stato
		c.State.Stage = 2

	// In questo stage è previsto che l'utenta debba effettuare una prima esplorazione
	case 2:
		// Invio messagio dove gli spiego che deve effettuare una nuova esplorazione
		firstUseMessage := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_use_item"),
		)
		firstUseMessage.ParseMode = "markdown"

		_, err = services.SendMessage(firstUseMessage)
		if err != nil {
			return
		}

		// Forzo a mano l'aggiornamento dello stato del player
		// in quanto adesso devo richiamare un'altro controller
		c.State.Stage = 3
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			return err
		}

		// Richiamo missione come sottoprocesso di questo controller
		useItemController := new(InventoryItemController)
		useItemController.Father = c.State.ID
		useItemController.Handle(c.Player, c.Update)

		// Recupero l'ID del task, mi serivirà per i controlli
		c.Payload.UseItemID = useItemController.State.ID

	// In questo stage è previsto che l'utenta debba effettuare una prima esplorazione
	case 3:
		// Invio messagio dove gli spiego che deve effettuare una nuova esplorazione
		firstMissionMessage := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_mission"),
		)
		firstMissionMessage.ParseMode = "markdown"

		_, err = services.SendMessage(firstMissionMessage)
		if err != nil {
			return
		}

		// Forzo a mano l'aggiornamento dello stato del player
		// in quanto adesso devo richiamare un'altro controller
		c.State.Stage = 5
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			return err
		}

		// Richiamo missione come sottoprocesso di questo controller
		missionController := new(MissionController)
		missionController.Father = c.State.ID
		missionController.Handle(c.Player, c.Update)

		// Recupero l'ID del task, mi serivirà per i controlli
		c.Payload.MissionID = missionController.State.ID
	// case 4:
	// 	// Primo Craft
	// 	_, err = services.SendMessage(
	// 		services.NewMessage(
	// 			c.Player.ChatID,
	// 			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_crafting"),
	// 		),
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	// Forzo a mano l'aggiornamento dello stato del player
	// 	// in quanto adesso devo richiamare un'altro controller
	// 	c.State.Stage = 5
	// 	c.State, err = providers.UpdatePlayerState(c.State)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	// Richiamo crafting come sottoprocesso di questo controller
	// 	craftingController := new(CraftingController)
	// 	craftingController.Father = c.State.ID
	// 	craftingController.Handle(c.Player, c.Update)
	//
	// 	// Recupero l'ID del task, mi serivirà per i controlli
	// 	c.Payload.CraftingID = craftingController.State.ID
	case 5:
		firstWeaponMessage := services.NewMessage(
			c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_weapon_equipped"),
		)
		firstWeaponMessage.ParseMode = "markdown"

		_, err = services.SendMessage(firstWeaponMessage)
		if err != nil {
			return err
		}

		// Forzo a mano l'aggiornamento dello stato del player
		// in quanto adesso devo richiamare un'altro controller
		c.State.Stage = 6
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			return err
		}

		// Richiamo crafting come sottoprocesso di questo controller
		inventoryController := new(InventoryEquipController)
		inventoryController.Father = c.State.ID
		inventoryController.Handle(c.Player, c.Update)

		// Recupero l'ID del task, mi serivirà per i controlli
		c.Payload.InventoryEquipID = inventoryController.State.ID

	case 6:
		firstHuntingMessage := services.NewMessage(
			c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_hunting"),
		)
		firstHuntingMessage.ParseMode = "markdown"

		_, err = services.SendMessage(firstHuntingMessage)
		if err != nil {
			return err
		}

		// Forzo a mano l'aggiornamento dello stato del player
		// in quanto adesso devo richiamare un'altro controller
		c.State.Stage = 7
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			return err
		}

		// Richiamo crafting come sottoprocesso di questo controller
		huntingController := new(HuntingController)
		huntingController.Father = c.State.ID
		huntingController.Handle(c.Player, c.Update)

		// Recupero l'ID del task, mi serivirà per i controlli
		c.Payload.HuntingID = huntingController.State.ID

	case 7:
		_, err = services.SendMessage(
			services.NewMessage(
				c.Player.ChatID,
				helpers.Trans(c.Player.Language.Slug, "route.tutorial.completed"),
			),
		)
		if err != nil {
			return err
		}

		// Addesso posso cancellare tutti gli stati associati
		_, err = providers.DeletePlayerState(nnsdk.PlayerState{ID: c.Payload.HuntingID})
		_, err = providers.DeletePlayerState(nnsdk.PlayerState{ID: c.Payload.InventoryEquipID})
		_, err = providers.DeletePlayerState(nnsdk.PlayerState{ID: c.Payload.CraftingID})
		_, err = providers.DeletePlayerState(nnsdk.PlayerState{ID: c.Payload.MissionID})
		if err != nil {
			return err
		}

		// Completo lo stato
		c.State.Completed = helpers.SetTrue()
	}

	return
}
