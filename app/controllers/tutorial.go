package controllers

import (
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
	Payload struct{}
}

// ====================================
// Handle
// ====================================
func (c *TutorialController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	// var isNewState bool
	c.Controller = "route.start"
	c.Player = player
	c.Update = update
	c.Message = update.Message

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	// Stato recuperto correttamente
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// log.Panicln(isNewState)

	// It's first message
	// if isNewState {
	// 	c.Stage()
	// 	return
	// }

	// Validate
	var hasError bool
	hasError, err = c.Validator()
	if err != nil {
		panic(err)
	}

	// Se ritornano degli errori
	if hasError == true {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
		services.SendMessage(validatorMsg)
		return
	}

	// Ok! Run!
	err = c.Stage()
	if err != nil {
		panic(err)
	}

	// Aggiorno stato finale
	_, err = providers.UpdatePlayerState(c.State)
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
		lang, err := providers.FindLanguageBy(c.Message.Text, "name")
		if err != nil {
			return false, err
		}

		// Verifico le la lingua esiste, se così non fosse ritorno errore
		if lang.ID <= 0 {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true, nil
		}

		return false, nil

	// In questo stage devo verificare unicamente che venga passata una stringa
	case 2:
		// Verifico che l'azione passata sia quella di aprire gli occhi
		if c.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.start.openEye") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			return true, nil
		}

		return false, nil

	// case 3:
	// c.Validation.Message = helpers.Trans("route.start.error.functionNotCompleted")
	// // Check if the player finished the previous function.
	// if c.State, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.mission"); c.State == (nnsdk.PlayerState{}) {
	// 	return false
	// }
	// case 4:
	// 	c.Validation.Message = helpers.Trans("route.start.error.functionNotCompleted")
	// 	// Check if the player finished the previous function.
	// 	if c.State, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.crafting"); c.State == (nnsdk.PlayerState{}) {
	// 		return false
	// 	}
	// case 5:
	// 	c.Validation.Message = helpers.Trans("route.start.error.functionNotCompleted")
	// 	// Check if the player finished the previous function.
	// 	if c.State, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.inventory.equip"); c.State == (nnsdk.PlayerState{}) {
	// 		return false
	// 	}
	// case 6:
	// 	c.Validation.Message = helpers.Trans("route.start.error.functionNotCompleted")
	// 	// Check if the player finished the previous function.
	// 	if c.State, _ = helpers.GetPlayerStateByFunction(helpers.Player, "route.hunting"); c.State == (nnsdk.PlayerState{}) {
	// 		return false
	// 	}
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
		msg := services.NewMessage(c.Message.Chat.ID, "Select language")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboard)
		services.SendMessage(msg)

		// Aggiorna stato
		c.State.Stage = 1

	// In questo stage è previsto un'invio di un set di messaggi
	// che introducono al player cosa sta accadendo
	case 1:
		// Recupero set di messaggi
		textList := helpers.GenerateTextArray(c.Player.Language.Slug, c.Controller)

		// Prendo il primo testo della intro e lo invio
		msg := services.NewMessage(c.Player.ChatID, textList[0])
		lastMessage := services.SendMessage(msg)

		// Mando primo set di messaggi
		var previousText string
		for i := 1; i < 3; i++ {
			time.Sleep(2 * time.Second)
			services.SendMessage(
				services.NewEditMessage(c.Player.ChatID, lastMessage.MessageID, textList[i]),
			)
		}

		// Invio altro set di messaggi
		for i := 3; i < 12; i++ {
			time.Sleep(2 * time.Second)
			previousText = services.SendMessage(
				services.NewEditMessage(c.Player.ChatID, lastMessage.MessageID, previousText+"\n"+textList[i]),
			).Text
		}

		lastMessage = services.SendMessage(services.NewMessage(c.Player.ChatID, textList[12]))
		previousText = lastMessage.Text
		for i := 13; i < len(textList); i++ {
			time.Sleep(time.Second)
			previousText = services.SendMessage(
				services.NewEditMessage(c.Player.ChatID, lastMessage.MessageID, previousText+"\n"+textList[i]),
			).Text
		}

		// Mando esplosione
		edit := services.NewEditMessage(
			c.Player.ChatID,
			lastMessage.MessageID,
			helpers.Trans(c.Player.Language.Slug, "route.start.explosion"),
		)
		edit.ParseMode = "HTML"
		services.SendMessage(edit)

		// Ultimo step apri gli occhi
		msg = services.NewMessage(c.Player.ChatID, "...")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.start.openEye")),
			),
		)
		services.SendMessage(msg)

		// Aggiorna stato
		c.State.Stage = 2

	// In questo stage è previsto che l'utenta debba effettuare una prima esplorazione
	case 2:
		// Invio messagio dove gli spiego che deve effettuare una nuova esplorazione
		services.SendMessage(
			services.NewMessage(
				c.Player.ChatID,
				helpers.Trans(c.Player.Language.Slug, "route.start.firstExploration"),
			),
		)

		// Forzo a mano l'aggiornamento dello stato del player
		// in quanto adesso devo richiamare un'altro controller
		c.State.Stage = 3
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			return err
		}

		// Richiamo missione come sottoprocesso di questo controller
		missionController := new(MissionController)
		missionController.Father = c.State.ID
		missionController.Handle(c.Player, c.Update)

		// case 3:
		// 	// First Crafting
		// 	services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstCrafting")))
		//
		// 	// Aggiorna stato
		// 	c.State.Stage = 4
		// 	c.State, err = providers.UpdatePlayerState(c.State)
		// 	if err != nil {
		// 		services.ErrorHandler("Cant update player", err)
		// 	}
		//
		// 	// Call crafting controller
		// 	new(CraftingController).Handle(c.Update)
		// case 4:
		// 	// Equip weapon
		// 	services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstWeaponEquipped")))
		//
		// 	// Aggiorna stato
		// 	c.State.Stage = 5
		// 	c.State, err = providers.UpdatePlayerState(c.State)
		// 	if err != nil {
		// 		services.ErrorHandler("Cant update player", err)
		// 	}
		//
		// 	// Call InventoryEquipController
		// 	new(InventoryEquipController).Handle(c.Update)
		// case 5:
		// 	services.SendMessage(services.NewMessage(helpers.Player.ChatID, helpers.Trans("route.start.firstHunting")))
		//
		// 	// Aggiorna stato
		// 	c.State.Stage = 6
		// 	c.State, err = providers.UpdatePlayerState(c.State)
		// 	if err != nil {
		// 		services.ErrorHandler("Cant update player", err)
		// 	}
		//
		// 	// Call InventoryEquipController
		// 	new(HuntingController).Handle(c.Update)
		// case 6:
		// 	//====================================
		// 	// COMPLETE!
		// 	//====================================
		// 	helpers.FinishAndCompleteState(c.State, helpers.Player)
		// 	//====================================
		// }
	}

	return
}
