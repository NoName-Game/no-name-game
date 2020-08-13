package controllers

import (
	"encoding/json"
	"strings"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// TitanPlanetTackleController
// ====================================
type TitanPlanetTackleController struct {
	BaseController
	Payload struct {
		CallbackChatID    int64
		CallbackMessageID int
		TitanID           uint32
		Selection         int32 // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		InFight           bool
		Kill              uint32
	}
}

// Settings generali
var (
	titanKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔼", "hunting.fight.up")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️", "hunting.fight.hit"),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", "hunting.fight.down")),
	)
)

// ====================================
// Handle
// ====================================
func (c *TitanPlanetTackleController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.titanplanet.tackle",
		c.Payload,
		[]string{},
		player,
		update,
	) {
		return
	}

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &MenuController{}) {
			return
		}
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.CurrentState.Payload, &c.Payload)

	// Validate
	var hasError bool
	hasError, err = c.Validator()
	if err != nil {
		panic(err)
	}

	// Se ritornano degli errori
	if hasError {
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
	if !hasError {
		err = c.Stage()
		if err != nil {
			panic(err)
		}
	}

	// Aggiorno stato finale
	payloadUpdated, _ := json.Marshal(c.Payload)
	c.CurrentState.Payload = string(payloadUpdated)

	var rUpdatePlayerState *pb.UpdatePlayerStateResponse
	rUpdatePlayerState, err = services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
		PlayerState: c.CurrentState,
	})
	if err != nil {
		panic(err)
	}
	c.CurrentState = rUpdatePlayerState.GetPlayerState()

	// Verifico completamento aggiuntivo per cancellare il messaggio
	if c.CurrentState.GetCompleted() {
		// Cancello messaggio contentente la mappa
		err = services.DeleteMessage(c.Payload.CallbackChatID, c.Payload.CallbackMessageID)
		if err != nil {
			panic(err)
		}
	}

	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *TitanPlanetTackleController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	// Il player deve avere sempre e perfoza un'arma equipaggiata
	// Indipendentemente dallo stato in cui si trovi
	if !helpers.CheckPlayerHaveOneEquippedWeapon(c.Player) {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.error.no_weapon_equipped")
		c.CurrentState.Completed = true
		return true, err
	}

	return false, err
}

// ====================================
// Stage
// ====================================
func (c *TitanPlanetTackleController) Stage() (err error) {
	switch c.CurrentState.Stage {
	// In questo stage faccio entrare il player nella mappa
	case 0:
		// Verifico se il player vuole uscire dalla caccia
		if c.Update.Message != nil {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.leave") {
				c.CurrentState.Completed = true
				return err
			}
		}

		// Avvio ufficialmente lo scontro!
		err = c.Tackle()
		if err != nil {
			return err
		}
	}

	return
}

// Hunting - in questo passo mi restituisco la mappa al player
func (c *TitanPlanetTackleController) Tackle() (err error) {
	// Recupero ultima posizione del player
	var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
	rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		return err
	}

	// Recupero titano in base alla posizione del player
	var rGetTitanByPlanetID *pb.GetTitanByPlanetIDResponse
	rGetTitanByPlanetID, err = services.NnSDK.GetTitanByPlanetID(helpers.NewContext(1), &pb.GetTitanByPlanetIDRequest{
		PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
	})

	if c.Update.Message != nil {
		// Se è qualsiasi messaggio diverso da affronta non lo calcolo
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.titanplanet.tackle") {
			return
		}

		// Questo messaggio è necessario per immettere il tasto di abbandona caccia
		initHunting := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.init"))
		initHunting.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.leave"),
				),
			),
		)
		_, err = services.SendMessage(initHunting)
		if err != nil {
			return err
		}

		//TODO: modificami
		test2 := helpers.Trans(c.Player.Language.Slug, "combat.card",
			rGetTitanByPlanetID.GetTitan().GetName(), strings.ToUpper("editmie"),
			rGetTitanByPlanetID.GetTitan().GetLifePoint(),
			rGetTitanByPlanetID.GetTitan().GetLifeMax(),
			c.Player.Username,
			c.PlayerStats.GetLifePoint(),
			100+c.PlayerStats.GetLevel()*10,
			helpers.Trans(c.Player.Language.Slug, bodyParts[c.Payload.Selection]),
		)

		// Invio quindi il mesaggio contenente mappa e azioni disponibili
		msg := services.NewMessage(c.Player.ChatID, test2)
		msg.ReplyMarkup = titanKeyboard
		msg.ParseMode = "HTML"

		var huntingMessage tgbotapi.Message
		huntingMessage, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno lo stato e ritorno
		// c.Payload.MapID = maps.ID
		c.Payload.CallbackChatID = huntingMessage.Chat.ID
		c.Payload.CallbackMessageID = huntingMessage.MessageID

		return err
	}

	// Se il messaggio è di tipo callback ed esiste una mappa associato al payload
	// potrebbe essere un messaggio lanciato da tasiterino, quindi acconsento allo spostamento
	if c.Update.CallbackQuery != nil {
		// Controllo tipo di callback data - move / fight
		actionType := strings.Split(c.Update.CallbackQuery.Data, ".")

		// Verifica tipo di movimento e mi assicuro che non sia in combattimento
		if actionType[1] == "fight" {
			err = c.Fight(actionType[2], rGetTitanByPlanetID.GetTitan())
		}

		if err != nil {
			return err
		}

		// Rimuove rotella di caricamento dal bottone
		err = services.AnswerCallbackQuery(
			services.NewAnswer(c.Update.CallbackQuery.ID, "", false),
		)

		return
	}

	return err
}

// ====================================
// Fight
// ====================================
func (c *TitanPlanetTackleController) Fight(action string, titan *pb.Titan) (err error) {
	var editMessage tgbotapi.EditMessageTextConfig

	switch action {
	// Avvio di un nuovo combattimento
	case "start":
		// Setto nuove informazioni stato
		c.Payload.TitanID = titan.GetID()
		c.Payload.InFight = true

	case "up":
		// Setto nuova parte del corpo da colpire
		if c.Payload.Selection > 0 {
			c.Payload.Selection--
		} else {
			c.Payload.Selection = 3
		}

	case "down":
		// Setto nuova parte del corpo da colpire
		if c.Payload.Selection < 3 {
			c.Payload.Selection++
		} else {
			c.Payload.Selection = 0
		}

	case "hit":
		// Effettuo chiamata al ws e recupero response dell'attacco
		var rHitTitan *pb.HitTitanResponse
		rHitTitan, err = services.NnSDK.HitTitan(helpers.NewContext(1), &pb.HitTitanRequest{
			TitanID:       titan.GetID(),
			PlayerID:      c.Player.ID,
			BodySelection: c.Payload.Selection,
		})
		if err != nil {
			return err
		}

		// Verifico se il MOB è morto
		if rHitTitan.GetTitanDie() {
			// Costruisco messaggio di recap del drop
			var dropRecap string

			// if rHitEnemy.GetEnemyDrop().GetResource() != nil {
			// 	dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.resource", rHitEnemy.GetEnemyDrop().GetResource().GetName())
			// } else if rHitEnemy.GetEnemyDrop().GetItem() != nil {
			// 	itemFound := helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", rHitEnemy.GetEnemyDrop().GetItem().GetSlug()))
			// 	dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.item", itemFound)
			// } else if rHitEnemy.GetEnemyDrop().GetTransaction() != nil {
			// 	dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.transaction", rHitEnemy.GetEnemyDrop().GetTransaction().GetValue())
			// } else {
			// 	dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.nothing")
			// }

			// Aggiungo anche esperinza recuperata
			// dropRecap += fmt.Sprintf("\n\n%s", helpers.Trans(c.Player.Language.Slug, "combat.experience", rHitEnemy.GetPlayerExperience()))

			// Aggiorno modifica del messaggio
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.mob_killed", titan.GetName(), dropRecap),
			)

			var ok = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						helpers.Trans(c.Player.Language.Slug, "continue"), "hunting.fight.return_map",
					),
				),
			)
			editMessage.ParseMode = "markdown"
			editMessage.ReplyMarkup = &ok

			// Setto stato
			c.Payload.Kill++
			c.Payload.InFight = false
			c.Payload.TitanID = 0

			// Invio messaggio
			_, err = services.SendMessage(editMessage)
			if err != nil {
				return err
			}

			return err
		}

		// Verifico se il PLAYER è morto
		if rHitTitan.GetPlayerDie() {
			// Aggiorno messaggio notificando al player che è morto
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.player_killed"),
			)

			var ok = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						helpers.Trans(c.Player.Language.Slug, "continue"), "hunting.fight.player-die",
					),
				),
			)

			editMessage.ReplyMarkup = &ok

			// Invio messaggio
			_, err = services.SendMessage(editMessage)
			if err != nil {
				return err
			}

			return err
		}

		// Se ne il player e ne il mob è morto, continua lo scontro
		// Messagio di notifica per vedere risultato attacco
		if rHitTitan.GetDodgeAttack() {
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.miss", rHitTitan.GetTitanDamage()),
			)
		} else {
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.damage", rHitTitan.GetPlayerDamage(), rHitTitan.GetTitanDamage()),
			)
		}

		ok := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action"),
			),
		)
		editMessage.ReplyMarkup = &ok
	case "player-die":
		// Il player è morto
		c.CurrentState.Completed = true

		return
	case "no-action":
		//
	}

	// Non sono state fatte modifiche al messaggio
	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
		editMessage = services.NewEditMessage(
			c.Player.ChatID,
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.Language.Slug, "combat.card",
				titan.GetName(), strings.ToUpper("editme"),
				titan.GetLifePoint(),
				titan.GetLifeMax(),
				c.Player.Username,
				c.PlayerStats.GetLifePoint(),
				100+c.PlayerStats.GetLevel()*10,
				helpers.Trans(c.Player.Language.Slug, bodyParts[c.Payload.Selection]),
			),
		)
		editMessage.ParseMode = "markdown"
		editMessage.ReplyMarkup = &titanKeyboard
	}

	// Invio messaggio modificato
	_, err = services.SendMessage(editMessage)

	return
}
