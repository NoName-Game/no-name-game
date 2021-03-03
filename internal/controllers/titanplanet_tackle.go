package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// TitanPlanetTackleController
// ====================================
type TitanPlanetTackleController struct {
	Controller
	Payload struct {
		CallbackMessageID int
		TitanID           uint32
		BodySelection     int32

		InEvent bool // Player have an event
		EventID uint32
	}
}

// Settings generali
var (
	fightTitanDie = helpers.InlineDataStruct{
		C:  "tackle",
		AT: "fight",
		A:  "titan_die",
	}

	titanFightKeyboard = [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔼", helpers.FightUp.GetDataString())),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️", helpers.FightHit.GetDataString()),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", helpers.FightDown.GetDataString())),
	}
)

// ====================================
// Handle
// ====================================
func (c *TitanPlanetTackleController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.titanplanet.tackle",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			PlanetType: []string{"titan"},
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
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *TitanPlanetTackleController) Validator() (hasErrors bool) {
	return false
}

// ====================================
// Stage
// ====================================
func (c *TitanPlanetTackleController) Stage() {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se il player vuole uscire dalla caccia
	// ##################################################################################################
	case 0:
		if c.Update.Message != nil {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.leave") {
				c.CurrentState.Completed = true

				if err := helpers.DeleteMessage(c.Player.ChatID, c.Payload.CallbackMessageID); err != nil {
					c.Logger.Panic(err)
				}

				return
			}
		}

		// Avvio ufficialmente lo scontro!
		c.Tackle()
	}

	return
}

// Tackle - Gestisco combattionmento con titano
func (c *TitanPlanetTackleController) Tackle() {
	var err error

	// Recupero posizione player corrente
	var playerPosition *pb.Planet
	if playerPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
		c.Logger.Panic(err)
	}

	// Recupero titano in base alla posizione del player
	var rGetTitanByPlanetID *pb.GetTitanByPlanetIDResponse
	if rGetTitanByPlanetID, err = config.App.Server.Connection.GetTitanByPlanetID(helpers.NewContext(1), &pb.GetTitanByPlanetIDRequest{
		PlanetID: playerPosition.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Se ricevo un messaggio normale probabilmente è un avvio o un abbandona
	if c.Update.Message != nil {
		// Se il titano è già stato ucciso esco
		if rGetTitanByPlanetID.GetTitan().GetKilledAt() != nil {
			// forzo l'uscita
			c.CurrentState.Completed = true
			return
		}
		// Se è qualsiasi messaggio diverso da affronta non lo calcolo
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.titanplanet.tackle") {
			return
		}

		// Questo messaggio è necessario per immettere il tasto di abbandona
		initHunting := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.init"))
		initHunting.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.leave"),
				),
			),
		)
		if _, err = helpers.SendMessage(initHunting); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero arma equipaggiata
		var rGetPlayerWeaponEquipped *pb.GetPlayerWeaponEquippedResponse
		rGetPlayerWeaponEquipped, _ = config.App.Server.Connection.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
			PlayerID: c.Player.ID,
		})

		weaponEquipped := helpers.Trans(c.Player.Language.Slug, "combat.no_weapon")
		var weaponDurability int32
		if rGetPlayerWeaponEquipped.GetWeapon().GetID() > 0 {
			weaponEquipped = rGetPlayerWeaponEquipped.GetWeapon().GetName()
			weaponDurability = rGetPlayerWeaponEquipped.GetWeapon().GetDurability()
		}

		// Preparo messaggio con la cardi di combattimento
		combactCard := helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.combat.card",
			rGetTitanByPlanetID.GetTitan().GetName(),
			rGetTitanByPlanetID.GetTitan().GetLifePoint(),
			rGetTitanByPlanetID.GetTitan().GetLifeMax(),
			c.Player.Username,
			c.Player.GetLifePoint(),
			c.Player.GetLevel().GetPlayerMaxLife(),
			helpers.Trans(c.Player.Language.Slug, bodyParts[c.Payload.BodySelection]), // Parte del corpo selezionata
			weaponEquipped, weaponDurability, // Arma equipaggiata e durabilità
		)

		// Invio quindi il mesaggio contenente le azioni disponibili
		msg := helpers.NewMessage(c.Player.ChatID, combactCard)

		// Inserisco fight keyboard
		if msg.ReplyMarkup, err = helpers.PlayerFightKeyboard(c.Player, titanFightKeyboard); err != nil {
			c.Logger.Panic(err)
		}

		var tackleMessage tgbotapi.Message
		msg.ParseMode = tgbotapi.ModeHTML
		if tackleMessage, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno lo stato e ritorno
		c.Payload.CallbackMessageID = tackleMessage.MessageID

		return
	}

	// Se il messaggio è di tipo callback sicuramete è un messaggio di attacco
	if c.Update.CallbackQuery != nil {
		// Controllo tipo di callback data - fight
		var inlineData helpers.InlineDataStruct
		inlineData = inlineData.GetDataValue(c.Update.CallbackQuery.Data)

		// Verifico che non sia in corso un'evento
		if c.Payload.InEvent && inlineData.AT == "event" {
			err = c.Event(inlineData, rGetTitanByPlanetID.GetTitan())
		} else if inlineData.AT == "fight" {
			// Verifica tipo di movimento e mi assicuro che non sia in combattimento
			err = c.Fight(inlineData, rGetTitanByPlanetID.GetTitan())
		}

		if err != nil {
			c.Logger.Panic(err)
		}

		// Rimuove rotella di caricamento dal bottone
		err = helpers.AnswerCallbackQuery(
			helpers.NewAnswer(c.Update.CallbackQuery.ID, "", false),
		)

		return
	}

	return
}

// ====================================
// Fight
// ====================================
func (c *TitanPlanetTackleController) Fight(inlineData helpers.InlineDataStruct, titan *pb.Titan) (err error) {
	switch inlineData.A {
	case "up":
		// Setto nuova parte del corpo da colpire
		if c.Payload.BodySelection > 0 {
			c.Payload.BodySelection--
		} else {
			c.Payload.BodySelection = 3
		}
	case "down":
		// Setto nuova parte del corpo da colpire
		if c.Payload.BodySelection < 3 {
			c.Payload.BodySelection++
		} else {
			c.Payload.BodySelection = 0
		}
	case "hit":
		c.Hit(titan, inlineData)
		return
	case "player_die":
		// Il player è morto
		c.CurrentState.Completed = true
		return
	case "titan_die":
		// Il titano è morto
		c.CurrentState.Completed = true
		return
	case "use":
		if err = helpers.UseItem(c.Player, inlineData.D, c.Payload.CallbackMessageID); err != nil {
			c.Logger.Panic(err)
		}
		return
	case "no_action":
		//
	}

	// Recupero arma equipaggiata
	var rGetPlayerWeaponEquipped *pb.GetPlayerWeaponEquippedResponse
	rGetPlayerWeaponEquipped, _ = config.App.Server.Connection.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
		PlayerID: c.Player.ID,
	})

	weaponEquipped := helpers.Trans(c.Player.Language.Slug, "combat.no_weapon")
	var weaponDurability int32
	if rGetPlayerWeaponEquipped.GetWeapon().GetID() > 0 {
		weaponEquipped = rGetPlayerWeaponEquipped.GetWeapon().GetName()
		weaponDurability = rGetPlayerWeaponEquipped.GetWeapon().GetDurability()
	}

	// Non sono state fatte modifiche al messaggio
	combactStatusMessage := helpers.NewEditMessage(
		c.Player.ChatID,
		c.Update.CallbackQuery.Message.MessageID,
		helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.combat.card",
			titan.GetName(),
			titan.GetLifePoint(),
			titan.GetLifeMax(),
			c.Player.Username,
			c.Player.GetLifePoint(),
			c.Player.GetLevel().GetPlayerMaxLife(),
			helpers.Trans(c.Player.Language.Slug, bodyParts[c.Payload.BodySelection]), // Parte del corpo selezionata
			weaponEquipped, weaponDurability, // Arma equipaggiata e durabilità
		),
	)

	// Inserisco fight keyboard
	if combactStatusMessage.ReplyMarkup, err = helpers.PlayerFightKeyboard(c.Player, titanFightKeyboard); err != nil {
		c.Logger.Panic(err)
	}

	combactStatusMessage.ParseMode = tgbotapi.ModeHTML
	if _, err = helpers.SendMessage(combactStatusMessage); err != nil {
		c.Logger.Panic(err)
	}

	return
}

func (c *TitanPlanetTackleController) Hit(titan *pb.Titan, inlineData helpers.InlineDataStruct) {
	var err error

	// Verifico se il player vuole usare un'abilità
	var abilityID uint32
	if inlineData.SA == "ability" {
		abilityID = inlineData.D
	}

	// Effettuo chiamata al ws e recupero response dell'attacco
	var rHitTitan *pb.HitTitanResponse
	if rHitTitan, err = config.App.Server.Connection.HitTitan(helpers.NewContext(1), &pb.HitTitanRequest{
		TitanID:       titan.GetID(),
		PlayerID:      c.Player.ID,
		BodySelection: c.Payload.BodySelection,
		AbilityID:     abilityID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Verifico se il TITANO è morto
	if rHitTitan.GetTitanDie() {
		c.TitanDie(titan)
		return
	}

	// Verifico se il PLAYER è morto
	if rHitTitan.GetPlayerDie() {
		c.PlayerDie(rHitTitan.GetTitanDamage())
		return
	}

	// Se ne il player e ne il mob è morto, continua lo scontro
	var combactMessage tgbotapi.EditMessageTextConfig
	if rHitTitan.GetDodgeAttack() {
		combactMessage = helpers.NewEditMessage(
			c.Player.ChatID,
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.Language.Slug, "combat.enemy_dodge", rHitTitan.GetTitanDamage()),
		)
	} else {
		combactMessage = helpers.NewEditMessage(
			c.Player.ChatID,
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.Language.Slug, "combat.damage", rHitTitan.GetPlayerDamage(), rHitTitan.GetTitanDamage()),
		)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Ok!", helpers.FightNoAction.GetDataString()),
		),
	)

	// Verifico se il prossimo colpo è un evento, se così fosse cambio keyborad
	if rHitTitan.GetTitanEventID() > 0 {
		c.Payload.InEvent = true
		c.Payload.EventID = rHitTitan.GetTitanEventID()

		// Aggiungo messggio che il titano si prepara ad un evento
		combactMessage.Text += helpers.Trans(c.Player.Language.Slug, "titanplanet.event.hungry")

		// Carico keyboard dedicata
		var dataEventStruct = helpers.InlineDataStruct{C: "titanplanet.tackle", AT: "event", A: "q"}
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(helpers.Trans(c.Player.Language.Slug, "titanplanet.event.get_ready"), dataEventStruct.GetDataString()),
			),
		)
	}

	combactMessage.ReplyMarkup = &keyboard
	combactMessage.ParseMode = tgbotapi.ModeHTML
	if _, err = helpers.SendMessage(combactMessage); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *TitanPlanetTackleController) PlayerDie(titanDamage int32) {
	// Aggiorno messaggio notificando al player che è morto
	playerDieMessage := helpers.NewEditMessage(
		c.Player.ChatID,
		c.Update.CallbackQuery.Message.MessageID,
		helpers.Trans(c.Player.Language.Slug, "combat.player_killed", titanDamage),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				helpers.Trans(c.Player.Language.Slug, "continue"), helpers.FightPlayerDie.GetDataString(),
			),
		),
	)

	playerDieMessage.ReplyMarkup = &keyboard
	playerDieMessage.ParseMode = tgbotapi.ModeHTML
	if _, err := helpers.SendMessage(playerDieMessage); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *TitanPlanetTackleController) TitanDie(titan *pb.Titan) {
	// Aggiorno modifica del messaggio
	titanDieMessage := helpers.NewEditMessage(
		c.Player.ChatID,
		c.Update.CallbackQuery.Message.MessageID,
		helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.combat.mob_killed", titan.GetName()),
	)

	var keyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				helpers.Trans(c.Player.Language.Slug, "continue"), fightTitanDie.GetDataString(),
			),
		),
	)

	titanDieMessage.ParseMode = tgbotapi.ModeHTML
	titanDieMessage.ReplyMarkup = &keyboard
	if _, err := helpers.SendMessage(titanDieMessage); err != nil {
		c.Logger.Panic(err)
	}

	// Setto stato
	c.Payload.TitanID = 0
}

// ====================================
// Event
// ====================================
func (c *TitanPlanetTackleController) Event(inlineData helpers.InlineDataStruct, titan *pb.Titan) (err error) {
	// Recupero evento in corso
	var rGetQuestion *pb.GetTitanEventQuestionByIDResponse
	if rGetQuestion, err = config.App.Server.Connection.GetTitanEventQuestionByID(helpers.NewContext(1), &pb.GetTitanEventQuestionByIDRequest{
		ID: c.Payload.EventID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	var question *pb.TitanEventQuestion
	question = rGetQuestion.GetQuestion()

	switch inlineData.A {
	case "q": // QUESTION
		// Recupero domande e risposte da dare al player
		questionMessage := helpers.NewEditMessage(
			c.Player.GetChatID(),
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.GetLanguage().GetSlug(), question.TextCode),
		)

		keyboardRow := new(tgbotapi.InlineKeyboardMarkup)
		for _, answer := range question.Answers {
			var dataAnswerStruct = helpers.InlineDataStruct{C: "titanplanet.tackle", AT: "event", A: "a", D: answer.ID}
			keyboardRow.InlineKeyboard = append(keyboardRow.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(helpers.Trans(c.Player.GetLanguage().GetSlug(), answer.GetTextCode()), dataAnswerStruct.GetDataString()),
			))
		}

		questionMessage.ReplyMarkup = keyboardRow
		if _, err := helpers.SendMessage(questionMessage); err != nil {
			c.Logger.Panic(err)
		}

	case "a": // ANSWER
		var rTitanEventSubmitAnswer *pb.TitanEventSubmitAnswerResponse
		if rTitanEventSubmitAnswer, err = config.App.Server.Connection.TitanEventSubmitAnswer(helpers.NewContext(1), &pb.TitanEventSubmitAnswerRequest{
			TitanID:  titan.ID,
			AnswerID: inlineData.D,
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Verifico se il TITANO è morto
		if rTitanEventSubmitAnswer.GetTitanDie() {
			c.TitanDie(titan)
			return
		}

		// Verifico se il PLAYER è morto
		if rTitanEventSubmitAnswer.GetPlayerDie() {
			c.PlayerDie(int32(rTitanEventSubmitAnswer.GetTitanDamage()))
			return
		}

		// Malus! | Bonus!
		var recap string
		if rTitanEventSubmitAnswer.GetIsMalus() {
			recap = helpers.Trans(c.Player.GetLanguage().GetSlug(), fmt.Sprintf(
				"titanplanet.event.question_%v.malus", rTitanEventSubmitAnswer.GetQuestionID(),
			), rTitanEventSubmitAnswer.GetTitanDamage())
		} else {
			recap = helpers.Trans(c.Player.GetLanguage().GetSlug(), fmt.Sprintf(
				"titanplanet.event.question_%v.bonus", rTitanEventSubmitAnswer.GetQuestionID(),
			), rTitanEventSubmitAnswer.GetPlayerDamage())
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(helpers.Trans(c.Player.GetLanguage().GetSlug(), "titanplanet.tackle.no_action"), helpers.FightNoAction.GetDataString()),
			),
		)

		recapMessage := helpers.NewEditMessage(c.Player.ChatID, c.Update.CallbackQuery.Message.MessageID, recap)
		recapMessage.ParseMode = tgbotapi.ModeHTML
		recapMessage.ReplyMarkup = &keyboard
		if _, err := helpers.SendMessage(recapMessage); err != nil {
			c.Logger.Panic(err)
		}

		c.Payload.InEvent = false
		c.Payload.EventID = 0
	}

	return
}
