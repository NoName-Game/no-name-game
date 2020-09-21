package controllers

// ====================================
// AbilityController
// ====================================
type AbilityController struct {
	Payload struct{}
	Controller
}

var (
	AbilityLists = []string{
		"Strength",
		// "dexterity",
		// "constitution",
		"Intellect",
		// "wisdom",
		// "charisma",
	}
)

// ====================================
// Handle
// ====================================
// func (c *AbilityController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
// 	// Inizializzo variabili del controler
// 	var err error
//
// 	// Verifico se è impossibile inizializzare
// 	if !c.InitController(
// 		"route.ability",
// 		c.Payload,
// 		[]string{},
// 		player,
// 		update,
// 	) {
// 		return
// 	}
//
// 	// Verifico se esistono condizioni per cambiare stato o uscire
// 	if !proxy {
// 		if c.BackTo(0, &MenuController{}) {
// 			return
// 		}
// 	}
//
// 	// Set and load payload
// 	helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)
//
// 	// Validate
// 	var hasError bool
// 	hasError, err = c.Validator()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// Se ritornano degli errori
// 	if hasError {
// 		// Invio il messaggio in caso di errore e chiudo
// 		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
// 		// validatorMsg.ReplyMarkup = c.Validation.ReplyKeyboard
//
// 		_, err = services.SendMessage(validatorMsg)
// 		if err != nil {
// 			panic(err)
// 		}
//
// 		return
// 	}
//
// 	// Ok! Run!
// 	err = c.Stage()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// Aggiorno stato finale
// 	payloadUpdated, _ := json.Marshal(c.Payload)
// 	c.PlayerData.CurrentState.Payload = string(payloadUpdated)
//
// 	rUpdatePlayerState, err := services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
// 		PlayerState: c.PlayerData.CurrentState,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	c.PlayerData.CurrentState = rUpdatePlayerState.GetPlayerState()
//
// 	// Verifico completamento
// 	err = c.Completing()
// 	if err != nil {
// 		panic(err)
// 	}
// }
//
// // ====================================
// // Validator
// // ====================================
// func (c *AbilityController) Validator() (hasErrors bool, err error) {
// 	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
// 	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
// 		tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton(
// 				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
// 			),
// 		),
// 	)
//
// 	switch c.PlayerData.CurrentState.Stage {
// 	// È il primo stato non c'è nessun controllo
// 	case 0:
// 		return false, err
//
// 	case 1:
// 		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "ability.back") {
// 			c.PlayerData.CurrentState.Stage = 0
// 			return false, err
// 		}
//
// 		// Verifico se l'abilità passata esiste nelle abilità censite e se il player ha punti disponibili
// 		for _, ability := range AbilityLists {
// 			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ability.%s", strings.ToLower(ability))) == c.Update.Message.Text {
// 				if c.PlayerData.PlayerStats.GetAbilityPoint() == 0 {
// 					c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ability.no_point_left")
// 					return true, err
// 				}
//
// 				return false, err
// 			}
// 		}
//
// 		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
// 		return true, err
// 	}
//
// 	return true, err
// }
//
// // ====================================
// // Stage
// // ====================================
// func (c *AbilityController) Stage() (err error) {
// 	switch c.PlayerData.CurrentState.Stage {
// 	// Invio messaggio con recap stats
// 	case 0:
// 		var recapStats string
// 		recapStats = helpers.Trans(c.Player.Language.Slug, "ability.stats.type")
//
// 		// Recupero dinamicamente i valory delle statistiche per poi ciclarli con quelli consentiti
// 		rv := reflect.ValueOf(&c.PlayerData.PlayerStats)
// 		rv = rv.Elem()
//
// 		for _, ability := range AbilityLists {
// 			playerStat := rv.FieldByName(ability)
// 			fieldName := helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ability.%s", strings.ToLower(ability)))
// 			recapStats += fmt.Sprintf("<code>%-15v:%v</code>\n", fieldName, playerStat)
// 		}
//
// 		// Mostro quanti punti ha a disposizione il player
// 		messagePlayerTotalPoint := helpers.Trans(c.Player.Language.Slug, "ability.stats.total_point", c.PlayerData.PlayerStats.GetAbilityPoint())
//
// 		// Creo tastierino con i soli componienti abilitati dal client
// 		var keyboardRow [][]tgbotapi.KeyboardButton
// 		for _, ability := range AbilityLists {
// 			row := tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(
// 					helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ability.%s", strings.ToLower(ability))),
// 				),
// 			)
// 			keyboardRow = append(keyboardRow, row)
// 		}
//
// 		// Aggiungo bottone cancella
// 		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton(
// 				helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
// 			),
// 		))
//
// 		msg := services.NewMessage(c.Player.ChatID, fmt.Sprintf("%s\n\n%s", messagePlayerTotalPoint, recapStats))
// 		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
// 			Keyboard:       keyboardRow,
// 			ResizeKeyboard: true,
// 		}
// 		msg.ParseMode = "HTML"
// 		_, err = services.SendMessage(msg)
// 		if err != nil {
// 			return err
// 		}
//
// 		// Avanzo di stage
// 		c.PlayerData.CurrentState.Stage = 1
// 	case 1:
// 		// Incremento statistiche e aggiorno
// 		for _, ability := range AbilityLists {
// 			abilityName := helpers.Trans(c.Player.Language.Slug, "ability."+strings.ToLower(ability))
//
// 			if abilityName == c.Update.Message.Text {
// 				f := reflect.ValueOf(&c.PlayerData.PlayerStats).Elem().FieldByName(ability)
// 				f.SetUint(uint64(f.Interface().(uint) + 1))
//
// 				c.PlayerData.PlayerStats.AbilityPoint--
// 			}
// 		}
//
// 		// TODO: Da rivedere in quanto bisognerebbe spostare la logica qui sopra
// 		// Aggiorno statistiche player
// 		// _, err = playerStatsProvider.UpdatePlayerStats(c.Player.Stats)
// 		// if err != nil {
// 		// 	return err
// 		// }
//
// 		// Invio Messaggio di incremento abilità
// 		text := helpers.Trans(c.Player.Language.Slug, "ability.stats.completed", c.Update.Message.Text)
// 		msg := services.NewMessage(c.Player.ChatID, text)
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
// 			),
// 		)
// 		_, err = services.SendMessage(msg)
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	return
// }
