package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// MissionController
// ====================================
type MissionController struct {
	BaseController
	Payload struct {
		ExplorationType string
		Times           int
		Material        nnsdk.Resource
		Quantity        int
	}
	// Additional Data
	MissionTypes []string
}

// ====================================
// Handle
// ====================================
func (c *MissionController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	// var isNewState bool
	c.Controller = "route.mission"
	c.Player = player
	c.Update = update
	c.Message = update.Message

	// Set mission types
	c.MissionTypes = make([]string, 3)
	c.MissionTypes[0] = helpers.Trans(c.Player.Language.Slug, "mission.underground")
	c.MissionTypes[1] = helpers.Trans(c.Player.Language.Slug, "mission.surface")
	c.MissionTypes[2] = helpers.Trans(c.Player.Language.Slug, "mission.atmosphere")

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	// Stato recuperto correttamente
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// Validate
	// var hasError bool
	// hasError, err = c.Validator()
	// if err != nil {
	// 	panic(err)
	// }
	//
	// // Se ritornano degli errori
	// if hasError == true {
	// 	// Invio il messaggio in caso di errore e chiudo
	// 	validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
	// 	services.SendMessage(validatorMsg)
	// 	return
	// }
	//
	// // Ok! Run!
	// err = c.Stage()
	// if err != nil {
	// 	panic(err)
	// }
	//
	// // Aggiorno stato finale
	// _, err = providers.UpdatePlayerState(c.State)
	// if err != nil {
	// 	panic(err)
	// }
	//
	// return
}

//
// //====================================
// // Validator
// //====================================
// func (c *MissionController) Validator() (hasErrors bool) {
// 	c.Validation.Message = helpers.Trans("validationMessage")
//
// 	switch c.State.Stage {
// 	case 1:
// 		// Controllo se il messaggio continene uno dei tipi di missione dichiarati
// 		if helpers.StringInSlice(c.Message.Text, c.MissionTypes) {
// 			c.Payload.ExplorationType = c.Message.Text
// 			return false
// 		}
// 	case 2:
// 		c.Validation.Message = helpers.Trans("mission.wait", c.State.FinishAt.Format("2006-01-02 15:04:05"))
// 		if time.Now().After(c.State.FinishAt) && c.Payload.Times < 10 {
// 			c.Payload.Times++
// 			c.Payload.Quantity = rand.Intn(3)*c.Payload.Times + 1
// 			return false
// 		}
// 	case 3:
// 		if c.Message.Text == helpers.Trans("mission.continue") {
// 			c.State.FinishAt = helpers.GetEndTime(0, 10*(2*c.Payload.Times), 0)
// 			c.State.ToNotify = helpers.SetTrue()
// 			return false
// 		} else if c.Message.Text == helpers.Trans("mission.comeback") {
// 			c.State.Stage = 4
// 			return false
// 		}
// 	}
//
// 	return true
// }
//
// //====================================
// // Stage
// //====================================
// func (c *MissionController) Stage() {
// 	var err error
//
// 	switch c.State.Stage {
// 	case 0:
// 		// Creo messaggio con la lista delle esplorazioni possibili
// 		var keyboardRows [][]tgbotapi.KeyboardButton
// 		for _, mType := range c.MissionTypes {
// 			keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(mType))
// 			keyboardRows = append(keyboardRows, keyboardRow)
// 		}
//
// 		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("mission.exploration"))
// 		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
// 			Keyboard:       keyboardRows,
// 			ResizeKeyboard: true,
// 		}
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		c.State.Stage = 1
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
// 	case 1:
// 		// Recupero materiali random
// 		material, err := providers.GetRandomResource(helpers.GetMissionCategoryID(c.Message.Text))
// 		if err != nil {
// 			services.ErrorHandler("Cant get random resources", err)
// 		}
//
// 		// Imposto nuovo stato
// 		c.Payload.Material = material
// 		jsonPayload, _ := json.Marshal(c.Payload)
// 		c.State.Payload = string(jsonPayload)
// 		c.State.Stage = 2
// 		c.State.ToNotify = helpers.SetTrue()
// 		c.State.FinishAt = helpers.GetEndTime(0, 10, 0)
//
// 		// Invio messaggio di attesa
// 		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("mission.wait", string(c.State.FinishAt.Format("15:04:05"))))
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("route.Menu"))))
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
//
// 		// Remove current redis state
// 		// helpers.DelRedisState(helpers.Player)
// 	case 2:
// 		// Invio messaggio di riepilogo con le materie recuperate e chiedo se vuole continuare o ritornare
// 		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("mission.extraction_recap", c.Payload.Material.Item.Name, c.Payload.Quantity))
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("mission.continue")),
// 				tgbotapi.NewKeyboardButton(helpers.Trans("mission.comeback")),
// 			),
// 		)
// 		services.SendMessage(msg)
//
// 		// Aggiorno lo stato
// 		jsonPayload, _ := json.Marshal(c.Payload)
// 		c.State.Payload = string(jsonPayload)
// 		c.State.Stage = 3
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
// 	case 3:
// 		// CONTINUE - Il player ha scelto di continuare la ricerca
// 		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("mission.wait", string(c.State.FinishAt.Format("15:04:05"))))
// 		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
// 		services.SendMessage(msg)
//
// 		// Aggiorno lo stato
// 		c.State.Stage = 2
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player", err)
// 		}
//
// 		// Remove current redist stare
// 		//helpers.DelRedisState(helpers.Player)
// 	case 4:
// 		// Invio messaggio di chiusura missione
// 		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("mission.extraction_ended"))
// 		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
// 		services.SendMessage(msg)
//
// 		// Aggiungo le risorse trovare dal player al suo inventario e chiudo
// 		_, err := providers.AddResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
// 			ItemID:   c.Payload.Material.ID,
// 			Quantity: c.Payload.Quantity,
// 		})
//
// 		if err != nil {
// 			services.ErrorHandler("Cant add resource to player inventory", err)
// 		}
//
// 		//====================================
// 		// COMPLETE!
// 		//====================================
// 		helpers.FinishAndCompleteState(c.State, helpers.Player)
// 		//====================================
// 	}
// }
