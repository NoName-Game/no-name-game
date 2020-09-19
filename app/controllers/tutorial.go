package controllers

import (
	"time"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/golang/protobuf/ptypes"
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
		UseItemID        uint32
		MissionID        uint32
		CraftingID       uint32
		HuntingID        uint32
		InventoryEquipID uint32
	}
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
		// c.BlockUpdateState = true
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

	// Set and load payload
	helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)

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
	if err = c.Completing(c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *TutorialController) Validator() (hasErrors bool) {
	var err error
	switch c.PlayerData.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	// In questo stage è necessario controllare se la lingua passata è quella giusta
	case 1:
		// Recupero lingue disponibili
		_, err = services.NnSDK.GetLanguageByName(helpers.NewContext(1), &pb.GetLanguageByNameRequest{
			Name: c.Update.Message.Text,
		})

		// Verifico se la lingua esiste, se così non fosse ritorno errore
		if err != nil {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true
		}

		return false

	// In questo stage devo verificare unicamente che venga passata una stringa
	case 2:
		// Verifico che l'azione passata sia quella di aprire gli occhi
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.tutorial.open_eye") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true
		}

		return false

	// In questo stage verifico se il player ha usato il rivitalizzante
	case 3:
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")

		var rGetPlayerStateByID *pb.GetPlayerStateByIDResponse
		rGetPlayerStateByID, err = services.NnSDK.GetPlayerStateByID(helpers.NewContext(1), &pb.GetPlayerStateByIDRequest{
			ID: c.Payload.UseItemID,
		})
		if err != nil {
			panic(err)
		}

		// Non è stato trovato lo stato ritorno allo stato precedente
		// e non ritorno errore
		if rGetPlayerStateByID.GetPlayerState().GetID() == 0 {
			c.PlayerData.CurrentState.Stage = 2
			return false
		}

		if !rGetPlayerStateByID.GetPlayerState().GetCompleted() {
			return true
		}

		return false

	// In questo stage verifico se il player ha completato correttamente la missione
	case 4:
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")

		var rGetPlayerStateByID *pb.GetPlayerStateByIDResponse
		rGetPlayerStateByID, err = services.NnSDK.GetPlayerStateByID(helpers.NewContext(1), &pb.GetPlayerStateByIDRequest{
			ID: c.Payload.MissionID,
		})
		if err != nil {
			panic(err)
		}

		// Non è stato trovato lo stato ritorno allo stato precedente
		// e non ritorno errore
		if rGetPlayerStateByID.GetPlayerState().GetID() == 0 {
			c.PlayerData.CurrentState.Stage = 3
			return false
		}

		if !rGetPlayerStateByID.GetPlayerState().GetCompleted() {
			return true
		}

		return false
	case 5:
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")

		var rGetPlayerStateByID *pb.GetPlayerStateByIDResponse
		rGetPlayerStateByID, err = services.NnSDK.GetPlayerStateByID(helpers.NewContext(1), &pb.GetPlayerStateByIDRequest{
			ID: c.Payload.InventoryEquipID,
		})
		if err != nil {
			panic(err)
		}

		// Non è stato trovato lo stato ritorno allo stato precedente
		// e non ritorno errore
		if rGetPlayerStateByID.GetPlayerState().GetID() == 0 {
			c.PlayerData.CurrentState.Stage = 4
			return false
		}

		if !rGetPlayerStateByID.GetPlayerState().GetCompleted() {
			return true
		}

		return false
	case 6:
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.tutorial.error.function_not_completed")

		var rGetPlayerStateByID *pb.GetPlayerStateByIDResponse
		rGetPlayerStateByID, err = services.NnSDK.GetPlayerStateByID(helpers.NewContext(1), &pb.GetPlayerStateByIDRequest{
			ID: c.Payload.HuntingID,
		})
		if err != nil {
			panic(err)
		}

		// Non è stato trovato lo stato ritorno allo stato precedente
		// e non ritorno errore
		if rGetPlayerStateByID.GetPlayerState().GetID() == 0 {
			c.PlayerData.CurrentState.Stage = 5
			return false
		}

		if !rGetPlayerStateByID.GetPlayerState().GetCompleted() {
			return true
		}

		return false
	case 7:
		var finishAt time.Time
		finishAt, err = ptypes.Timestamp(c.PlayerData.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"ship.travel.wait",
			finishAt.Format("15:04:05 01/02"),
		)

		// Verifico se ha finito il crafting
		if time.Now().After(finishAt) {
			return false
		}

		return true
	case 8:
		return false
	default:
		// Stato non riconosciuto ritorno errore
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.state")
	}

	return true
}

// ====================================
// Stage - Language -> Messages -> Exploration -> Crafting -> Hunting
// ====================================
func (c *TutorialController) Stage() (err error) {
	switch c.PlayerData.CurrentState.Stage {
	// Primo avvio in questo momento l'utente deve poter ricevere la lista delle lingue disponibili
	// e potrà selezionare la sua lingua tramite tastierino
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
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorna stato
		c.PlayerData.CurrentState.Stage = 1

	// #############################
	// SET MESSAGGI
	// #############################
	case 1:
		// Recupero set di messaggi
		// textList := helpers.GenerateTextArray(c.Player.Language.Slug, c.Configuration.Controller)
		//
		// // Invio il primo messaggio per eliminare la tastiera
		// initMessage := services.NewMessage(c.Update.Message.Chat.ID, "...")
		// initMessage.ParseMode = "markdown"
		// initMessage.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		// _, err = services.SendMessage(initMessage)
		// if err != nil {
		// 	return err
		// }
		//
		// // ************************
		// // Primo set di messaggi
		// // ************************
		// time.Sleep(1 * time.Second)
		//
		// var firstMessageConfig tgbotapi.MessageConfig
		// firstMessageConfig = services.NewMessage(c.Player.ChatID, textList[0])
		// firstMessageConfig.ParseMode = "markdown"
		//
		// var firstMessage tgbotapi.Message
		// firstMessage, err = services.SendMessage(firstMessageConfig)
		// if err != nil {
		// 	return err
		// }
		//
		// // Mando primo set di messaggi
		// for i := 1; i <= 7; i++ {
		// 	time.Sleep(1 * time.Second)
		// 	edited := services.NewEditMessage(c.Player.ChatID, firstMessage.MessageID, textList[i])
		// 	edited.ParseMode = "markdown"
		//
		// 	_, err = services.SendMessage(edited)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		//
		// // ************************
		// // Secondo set di messaggi
		// // ************************
		// time.Sleep(1 * time.Second)
		// var secondSetText = textList[8]
		//
		// var secondMessageConfig tgbotapi.MessageConfig
		// secondMessageConfig = services.NewMessage(c.Player.ChatID, secondSetText)
		// secondMessageConfig.ParseMode = "markdown"
		//
		// var secondMessage tgbotapi.Message
		// secondMessage, err = services.SendMessage(secondMessageConfig)
		// if err != nil {
		// 	return err
		// }
		//
		// // PreviusText mi serve per andare a modificare il messaggio
		// // inviato ed appendergli la nuova parte di messaggio
		// for i := 9; i <= 12; i++ {
		// 	time.Sleep(2 * time.Second)
		// 	currentMessage := fmt.Sprintf("%s%s", secondSetText, textList[i])
		//
		// 	edited := services.NewEditMessage(
		// 		c.Player.ChatID,
		// 		secondMessage.MessageID,
		// 		currentMessage,
		// 	)
		// 	edited.ParseMode = "markdown"
		//
		// 	secondMessage, err = services.SendMessage(edited)
		// 	if err != nil {
		// 		return
		// 	}
		//
		// 	// Concateno messaggi
		// 	secondSetText += textList[i]
		// }
		//
		// // ************************
		// // Terzo set di messaggi
		// // ************************
		// time.Sleep(1 * time.Second)
		// thirdSetText := textList[13]
		//
		// var thirdMessageConfig tgbotapi.MessageConfig
		// thirdMessageConfig = services.NewMessage(c.Player.ChatID, thirdSetText)
		// thirdMessageConfig.ParseMode = "markdown"
		//
		// var thirdMessage tgbotapi.Message
		// thirdMessage, err = services.SendMessage(thirdMessageConfig)
		// if err != nil {
		// 	return err
		// }
		//
		// // PreviusText mi serve per andare a modificare il messaggio
		// // inviato ed appendergli la nuova parte di messaggio
		// for i := 14; i <= 19; i++ {
		// 	currentMessage := fmt.Sprintf("%s%s", thirdSetText, textList[i])
		//
		// 	time.Sleep(2 * time.Second)
		// 	edited := services.NewEditMessage(
		// 		c.Player.ChatID,
		// 		thirdMessage.MessageID,
		// 		currentMessage,
		// 	)
		// 	edited.ParseMode = "markdown"
		//
		// 	thirdMessage, err = services.SendMessage(edited)
		// 	if err != nil {
		// 		return
		// 	}
		//
		// 	thirdSetText += textList[i]
		// }
		//
		// // Mando messaggio allert
		// time.Sleep(2 * time.Second)
		// alertMessage := services.NewMessage(c.Update.Message.Chat.ID, textList[20])
		// alertMessage.ParseMode = "markdown"
		// alertMessage.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		// _, err = services.SendMessage(alertMessage)
		// if err != nil {
		// 	return err
		// }
		//
		// // ************************
		// // Quarto set di messaggi ( COUNTDOWN )
		// // ************************
		// time.Sleep(2 * time.Second)
		// var fourthMessageConfig tgbotapi.MessageConfig
		// fourthMessageConfig = services.NewMessage(c.Player.ChatID, textList[21])
		// fourthMessageConfig.ParseMode = "markdown"
		//
		// var fourthMessage tgbotapi.Message
		// fourthMessage, err = services.SendMessage(fourthMessageConfig)
		// if err != nil {
		// 	return err
		// }
		//
		// // Mando primo set di messaggi
		// for i := 22; i <= 27; i++ {
		// 	time.Sleep(1 * time.Second)
		// 	edited := services.NewEditMessage(
		// 		c.Player.ChatID,
		// 		fourthMessage.MessageID,
		// 		textList[i],
		// 	)
		//
		// 	edited.ParseMode = "markdown"
		// 	_, err = services.SendMessage(edited)
		// 	if err != nil {
		// 		return err
		// 	}
		// }
		//
		// // ************************
		// // Esplosione
		// // ************************
		// edit := services.NewEditMessage(
		// 	c.Player.ChatID,
		// 	fourthMessage.MessageID,
		// 	helpers.Trans(c.Player.Language.Slug, "route.tutorial.explosion"),
		// )
		//
		// edit.ParseMode = "HTML"
		// _, err = services.SendMessage(edit)
		// if err != nil {
		// 	return
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
		_, err = services.SendMessage(openEyeMessage)
		if err != nil {
			return
		}

		// Aggiorna stato
		c.PlayerData.CurrentState.Stage = 2

	// #############################
	// USA ITEM
	// #############################
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
		c.PlayerData.CurrentState.Stage = 3

		var rUpdatePlayerState *pb.UpdatePlayerStateResponse
		rUpdatePlayerState, err = services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
			PlayerState: c.PlayerData.CurrentState,
		})
		if err != nil {
			return err
		}

		c.PlayerData.CurrentState = rUpdatePlayerState.GetPlayerState()

		// Richiamo missione come sottoprocesso di questo controller
		useItemController := new(InventoryItemController)
		useItemController.ControllerFather = c.PlayerData.CurrentState.ID
		useItemController.Handle(c.Player, c.Update)

		// Recupero l'ID del task, mi serivirà per i controlli
		c.Payload.UseItemID = useItemController.PlayerData.CurrentState.ID

	// #############################
	// PRIMA ESPLORAZIONE
	// #############################
	case 3:
		// Invio messagio dove gli spiego che deve effettuare una nuova esplorazione
		firstMissionMessage := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_exploration"),
		)
		firstMissionMessage.ParseMode = "markdown"

		_, err = services.SendMessage(firstMissionMessage)
		if err != nil {
			return
		}

		// Forzo a mano l'aggiornamento dello stato del player
		// in quanto adesso devo richiamare un'altro controller
		c.PlayerData.CurrentState.Stage = 4

		var rUpdatePlayerState *pb.UpdatePlayerStateResponse
		rUpdatePlayerState, err = services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
			PlayerState: c.PlayerData.CurrentState,
		})
		if err != nil {
			return
		}

		c.PlayerData.CurrentState = rUpdatePlayerState.GetPlayerState()

		// Richiamo missione come sottoprocesso di questo controller
		missionController := new(ExplorationController)
		missionController.ControllerFather = c.PlayerData.CurrentState.ID
		// missionController.Payload.ForcedTime = 1
		missionController.Handle(c.Player, c.Update)

		// Recupero l'ID del task, mi serivirà per i controlli
		c.Payload.MissionID = missionController.PlayerData.CurrentState.ID

	// #############################
	// EQUIPAGGIA ARMA
	// #############################
	case 4:
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
		c.PlayerData.CurrentState.Stage = 5

		var rUpdatePlayerState *pb.UpdatePlayerStateResponse
		rUpdatePlayerState, err = services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
			PlayerState: c.PlayerData.CurrentState,
		})
		if err != nil {
			return err
		}

		c.PlayerData.CurrentState = rUpdatePlayerState.GetPlayerState()

		// Richiamo crafting come sottoprocesso di questo controller
		inventoryController := new(PlayerEquipmentController)
		inventoryController.ControllerFather = c.PlayerData.CurrentState.ID
		inventoryController.Handle(c.Player, c.Update)

		// Recupero l'ID del task, mi serivirà per i controlli
		c.Payload.InventoryEquipID = inventoryController.PlayerData.CurrentState.ID

	// #############################
	// CACCIA!
	// #############################
	case 5:
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
		c.PlayerData.CurrentState.Stage = 6

		var rUpdatePlayerState *pb.UpdatePlayerStateResponse
		rUpdatePlayerState, err = services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
			PlayerState: c.PlayerData.CurrentState,
		})
		if err != nil {
			return err
		}

		c.PlayerData.CurrentState = rUpdatePlayerState.GetPlayerState()

		// Richiamo crafting come sottoprocesso di questo controller
		huntingController := new(HuntingController)
		huntingController.ControllerFather = c.PlayerData.CurrentState.ID
		huntingController.Handle(c.Player, c.Update)

		// Recupero l'ID del task, mi serivirà per i controlli
		c.Payload.HuntingID = huntingController.PlayerData.CurrentState.ID

	// #############################
	// PRIMO VIAGGIO VERSO PIANETA SICURO
	// #############################
	case 6:
		// Questo stage fa viaggiare il player forzatamente verso un pianeta sicuro
		firstTravelMessage := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_travel"))
		firstTravelMessage.ParseMode = "markdown"
		_, err = services.SendMessage(firstTravelMessage)
		if err != nil {
			return err
		}

		finishTime := helpers.GetEndTime(0, 30, 0)
		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.travel.exploring", finishTime.Format("15:04:05 01/02")),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Forzo a mano l'aggiornamento dello stato del player
		// in quanto adesso devo richiamare un'altro controller
		c.PlayerData.CurrentState.Stage = 7
		c.PlayerData.CurrentState.FinishAt, _ = ptypes.TimestampProto(finishTime)
		c.ForceBackTo = true
	case 7:
		// var rGetShipEquipped *pb.GetPlayerShipEquippedResponse
		// rGetShipEquipped, err = services.NnSDK.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
		// 	PlayerID: c.Player.ID,
		// })
		// if err != nil {
		// 	return err
		// }
		//
		// // Recupero la posizione del player e i pianeti sicuro
		// var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
		// rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
		// 	PlayerID: c.Player.ID,
		// })
		// if err != nil {
		// 	return err
		// }
		//
		// systemID := rGetPlayerCurrentPlanet.GetPlanet().GetPlanetSystemID()
		//
		// var rGetSafePlanet *pb.GetSafePlanetsResponse
		// rGetSafePlanet, err = services.NnSDK.GetSafePlanets(helpers.NewContext(1), &pb.GetSafePlanetsRequest{})
		// if err != nil {
		// 	return err
		// }
		//
		// var safePlanet *pb.Planet
		// for _, p := range rGetSafePlanet.GetSafePlanets() {
		// 	if p.GetPlanetSystemID() == systemID {
		// 		// Il pianeta sicuro è quello del sistema del player
		// 		safePlanet = p
		// 	}
		// }
		//
		// _, err = services.NnSDK.EndShipTravel(helpers.NewContext(1), &pb.EndShipTravelRequest{
		// 	PlayerID: c.Player.ID,
		// 	// Integrity: 0,
		// 	// Tank:      0,
		// 	// PlanetID:  safePlanet.ID,
		// 	// ShipID:    rGetShipEquipped.GetShip().GetID(),
		// })
		// if err != nil {
		// 	return err
		// }

		firstSafeMessage := services.NewMessage(
			c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "route.tutorial.first_safeplanet"),
		)
		firstSafeMessage.ParseMode = "markdown"

		_, err = services.SendMessage(firstSafeMessage)
		if err != nil {
			return err
		}
		// Forzo a mano l'aggiornamento dello stato del player
		// in quanto adesso devo richiamare un'altro controller
		c.PlayerData.CurrentState.Stage = 8
		c.ForceBackTo = true
	case 8:
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
		var playerStatesIDToDelete = []uint32{
			c.Payload.UseItemID,
			c.Payload.HuntingID,
			c.Payload.InventoryEquipID,
			c.Payload.CraftingID,
			c.Payload.MissionID,
		}

		for _, stateID := range playerStatesIDToDelete {
			_, err = services.NnSDK.DeletePlayerState(helpers.NewContext(1), &pb.DeletePlayerStateRequest{
				PlayerStateID: stateID,
				ForceDelete:   true,
			})
			if err != nil {
				return err
			}
		}

		// Registro che il player ha completato il tutorial e recupero rewward
		var rPlayerEndTutorial *pb.PlayerEndTutorialResponse
		rPlayerEndTutorial, err = services.NnSDK.PlayerEndTutorial(helpers.NewContext(1), &pb.PlayerEndTutorialRequest{
			PlayerID: c.Player.ID,
			End:      true,
		})
		if err != nil {
			return err
		}

		rewardMessage := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
			"route.tutorial.completed.reward",
			rPlayerEndTutorial.GetMoney(),
			rPlayerEndTutorial.GetExp(),
		))
		rewardMessage.ParseMode = "markdown"

		_, err = services.SendMessage(rewardMessage)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.PlayerData.CurrentState.Completed = true
	}

	return
}
