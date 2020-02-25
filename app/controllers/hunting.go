package controllers

import (
	"encoding/json"
	"errors"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// HuntingController
//
// In questo controller il player avrà la possibilità di esplorare
// la mappa del pianeta che sta visitando, e di conseguenza affrontare mob,
// recupeare tesori e cascare in delle trappole
// ====================================
type HuntingController struct {
	BaseController
	Payload struct {
		CallbackChatID    int64
		CallbackMessageID int
		MapID             uint
		EnemyID           uint
		Selection         int // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		InFight           bool
		Kill              uint
	}
	PlayerPositionX int
	PlayerPositionY int
}

// Settings generali
var (
	// Antiflood
	antiFloodSeconds float64 = 1.0

	// Parti di corpo disponibili per l'attacco
	bodyParts = [4]string{"head", "chest", "gauntlets", "leg"}

	// Keyboard inline di esplorazione
	mapKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", "hunting.move.up")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", "hunting.move.left"),
			tgbotapi.NewInlineKeyboardButtonData("⭕", "hunting.move.action"),
			tgbotapi.NewInlineKeyboardButtonData("➡️", "hunting.move.right"),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", "hunting.move.down")),
	)

	fightKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", "hunting.move.up")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", "hunting.move.left"),
			tgbotapi.NewInlineKeyboardButtonData("⚔️", "hunting.fight.start"),
			tgbotapi.NewInlineKeyboardButtonData("➡️", "hunting.move.right"),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", "hunting.move.down")),
	)

	mobKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔼", "hunting.fight.up")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⚔", "hunting.fight.hit")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", "hunting.fight.down")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🗾", "hunting.fight.return_map")),
	)
)

// ====================================
// Handle
// ====================================
func (c *HuntingController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	c.Controller = "route.hunting"
	c.Player = player
	c.Update = update

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	// Set and load payload
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
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
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
	c.State, err = providers.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico completamento aggiuntivo per cancellare il messaggio
	if *c.State.Completed == true {
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

	return
}

// ====================================
// Validator
// ====================================
func (c *HuntingController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	// Il player deve avere sempre e perfoza un'arma equipaggiata
	// Indipendentemente dallo stato in cui si trovi
	if !helpers.CheckPlayerHaveOneEquippedWeapon(c.Player) {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "hunting.error.no_weapon_equipped")

		return true, err
	}

	return false, err
}

// ====================================
// Stage Map -> Drop -> Finish
// ====================================
func (c *HuntingController) Stage() (err error) {
	switch c.State.Stage {
	// In questo stage faccio entrare il player nella mappa
	case 0:
		// Verifico se il player vuole uscire dalla caccia
		if c.Update.Message != nil {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "hunting.leave") {
				c.State.Completed = helpers.SetTrue()
				return err
			}
		}

		// Avvio ufficialmente la caccia!
		err = c.Hunting()
		if err != nil {
			return err
		}
	}

	return
}

// Hunting - in questo passo mi restituisco la mappa al player
func (c *HuntingController) Hunting() (err error) {
	// Se nel payload NON è presente un ID della mappa lo
	// recupero dalla posizione del player e invio al player il messaggio
	// principale contenente la mappa e il tastierino
	if c.Payload.MapID <= 0 || c.Update.Message != nil {
		// Questo messaggio è necessario per immettere il tasto di abbandona caccia
		initHunting := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "hunting.init"))
		initHunting.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "hunting.leave"),
				),
			),
		)
		_, err = services.SendMessage(initHunting)
		if err != nil {
			return err
		}

		// Recupero ultima posizione del player, dando per scontato che sia
		// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
		var lastPosition nnsdk.PlayerPosition
		lastPosition, err = providers.GetPlayerLastPosition(c.Player)
		if err != nil {
			return err
		}

		// Dalla ultima posizione recupero il pianeta corrente
		var planet nnsdk.Planet
		planet, err = providers.GetPlanetByCoordinate(lastPosition.X, lastPosition.Y, lastPosition.Z)
		if err != nil {
			return err
		}

		// Recupero dettagli della mappa e per non appesantire le chiamate
		// al DB registro il tutto su redis
		var maps nnsdk.Map
		maps, err = providers.GetMapByID(planet.Map.ID)
		if err != nil {
			return err
		}

		// Registro mappa e posizione iniziale del player
		err = helpers.SetRedisMapHunting(maps)
		err = helpers.SetRedisPlayerHuntingPosition(maps, c.Player, "X", maps.StartPositionX)
		err = helpers.SetRedisPlayerHuntingPosition(maps, c.Player, "Y", maps.StartPositionY)
		if err != nil {
			return err
		}

		// Trasformo la mappa in qualcosa di più leggibile su telegram
		var decodedMap string
		decodedMap, err = helpers.DecodeMapToDisplay(maps, maps.StartPositionX, maps.StartPositionY)
		if err != nil {
			return err
		}

		// Invio quindi il mesaggio contenente mappa e azioni disponibili
		msg := services.NewMessage(c.Player.ChatID, decodedMap)
		msg.ReplyMarkup = mapKeyboard
		msg.ParseMode = "HTML"

		var huntingMessage tgbotapi.Message
		huntingMessage, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno lo stato e ritorno
		c.Payload.MapID = maps.ID
		c.Payload.CallbackChatID = huntingMessage.Chat.ID
		c.Payload.CallbackMessageID = huntingMessage.MessageID
		return
	}

	// Se il messaggio è di tipo callback ed esiste una mappa associato al payload
	// potrebbe essere un messaggio lanciato da tasiterino, quindi acconsento allo spostamento
	if c.Payload.MapID > 0 && c.Update.CallbackQuery != nil {
		var maps nnsdk.Map
		maps, err = helpers.GetRedisMapHunting(c.Payload.MapID)
		if err != nil {
			return err
		}

		// Recupero posizione player
		// var playerPositionX, playerPositionY int
		c.PlayerPositionX, err = helpers.GetRedisPlayerHuntingPosition(maps, c.Player, "X")
		c.PlayerPositionY, err = helpers.GetRedisPlayerHuntingPosition(maps, c.Player, "Y")

		// Controllo tipo di callback data - move / fight
		actionType := strings.Split(c.Update.CallbackQuery.Data, ".")

		// Verifica tipo di movimento e mi assicuro che non sia in combattimento
		if actionType[1] == "move" && !c.Payload.InFight {
			err = c.Move(actionType[2], maps)
		} else if actionType[1] == "fight" {
			err = c.Fight(actionType[2], maps)
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
// Movements
// ====================================
func (c *HuntingController) Move(action string, maps nnsdk.Map) (err error) {
	// Refresh della mappa
	var cellGrid [][]bool
	err = json.Unmarshal([]byte(maps.CellGrid), &cellGrid)
	if err != nil {
		return err
	}

	// Do scontato che sia una quadrato
	dimension := len(cellGrid)

	// Eseguo azione
	switch action {
	case "up":
		// Verifico limiti mappa
		if c.PlayerPositionX-1 < 0 {
			c.PlayerPositionX++
			break
		}

		if !cellGrid[c.PlayerPositionX-1][c.PlayerPositionY] {
			c.PlayerPositionX--
		} else {
			c.PlayerPositionX++
		}
	case "down":
		// Verifico limiti mappa
		if c.PlayerPositionX+1 >= dimension {
			c.PlayerPositionX--
			break
		}

		if !cellGrid[c.PlayerPositionX+1][c.PlayerPositionY] {
			c.PlayerPositionX++
		} else {
			c.PlayerPositionX--
		}
	case "left":
		// Verifico limiti mappa
		if c.PlayerPositionY-1 < 0 {
			c.PlayerPositionY++
			break
		}

		if !cellGrid[c.PlayerPositionX][c.PlayerPositionY-1] {
			c.PlayerPositionY--
		} else {
			c.PlayerPositionY++
		}
	case "right":
		// Verifico limiti mappa
		if c.PlayerPositionY+1 >= dimension {
			c.PlayerPositionY--
			break
		}

		if !cellGrid[c.PlayerPositionX][c.PlayerPositionY+1] {
			c.PlayerPositionY++
		} else {
			c.PlayerPositionY--
		}
	case "action":
		// Verifico se si trova sopra un tesoro se così fosse lancio
		// chiamata per verificare il drop
		var nearTresure bool
		var tresure nnsdk.Tresure
		tresure, nearTresure = helpers.CheckForTresure(maps, c.PlayerPositionX, c.PlayerPositionY)
		if nearTresure == true {
			// Chiamo WS e recupero tesoro
			var drop nnsdk.TresureDropResponse
			drop, err = providers.DropTresure(c.Player.ID, tresure.ID)
			if err != nil {
				return err
			}

			// Verifico cosa è tornato e rispondo
			var editMessage tgbotapi.EditMessageTextConfig
			if drop.Resource.ID > 0 {

				editMessage = services.NewEditMessage(
					c.Player.ChatID,
					c.Update.CallbackQuery.Message.MessageID,
					helpers.Trans(c.Player.Language.Slug, "tresure.found.resource", drop.Resource.Name),
				)

			} else if drop.Item.ID > 0 {

				editMessage = services.NewEditMessage(
					c.Player.ChatID,
					c.Update.CallbackQuery.Message.MessageID,
					helpers.Trans(c.Player.Language.Slug, "tresure.found.item", drop.Item.Name),
				)

			} else if drop.Transaction.ID > 0 {

				editMessage = services.NewEditMessage(
					c.Player.ChatID,
					c.Update.CallbackQuery.Message.MessageID,
					helpers.Trans(c.Player.Language.Slug, "tresure.found.transaction", drop.Transaction.Value),
				)

			} else {
				editMessage = services.NewEditMessage(
					c.Player.ChatID,
					c.Update.CallbackQuery.Message.MessageID,
					helpers.Trans(c.Player.Language.Slug, "tresure.found.nothing"),
				)
			}

			ok := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.move.no-action"),
				),
			)
			editMessage.ReplyMarkup = &ok
			editMessage.ParseMode = "HTML"

			_, err = services.SendMessage(editMessage)
			if err != nil {
				return err
			}

			return
		}
	case "no-action":
		// No action
	default:
		err = errors.New("action not recognized")
		return err
	}

	// Aggiorno nuova posizione del player
	err = helpers.SetRedisPlayerHuntingPosition(maps, c.Player, "X", c.PlayerPositionX)
	err = helpers.SetRedisPlayerHuntingPosition(maps, c.Player, "Y", c.PlayerPositionY)
	if err != nil {
		return
	}

	// Trasformo la mappa in qualcosa di più leggibile su telegram
	var decodedMap string
	decodedMap, err = helpers.DecodeMapToDisplay(maps, c.PlayerPositionX, c.PlayerPositionY)
	if err != nil {
		return err
	}

	// Se l'azione è valida e completa aggiorno risultato
	msg := services.NewEditMessage(c.Player.ChatID, c.Update.CallbackQuery.Message.MessageID, decodedMap)

	// Se un player si trova sulla stessa posizione un mob o di un tesoro effettuo il controllo
	var nearMob bool
	_, nearMob = helpers.CheckForMob(maps, c.PlayerPositionX, c.PlayerPositionY)
	if true == nearMob {
		msg.ReplyMarkup = &fightKeyboard
	} else {
		msg.ReplyMarkup = &mapKeyboard
	}

	msg.ParseMode = "HTML"
	_, err = services.SendMessage(msg)
	if err != nil {
		return err
	}

	return
}

// ====================================
// Fight
// ====================================
func (c *HuntingController) Fight(action string, maps nnsdk.Map) (err error) {
	var enemy nnsdk.Enemy
	var editMessage tgbotapi.EditMessageTextConfig

	if c.Payload.EnemyID > 0 {
		// Se impostato recupero informazioni più aggiornate del mob
		enemy, _ = providers.GetEnemyByID(c.Payload.EnemyID)
	} else {
		// Recupero il mob più vicino con il quale combattere e me lo setto nel payload
		enemy, _ = helpers.CheckForMob(maps, c.PlayerPositionX, c.PlayerPositionY)
	}

	switch action {
	// Avvio di un nuovo combattimento
	case "start":
		// Setto nuove informazioni stato
		c.Payload.EnemyID = enemy.ID
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
		var hitRequest nnsdk.HitEnemyRequest
		hitRequest = nnsdk.HitEnemyRequest{
			PlayerID:        c.Player.ID,
			PlayerPositionX: c.PlayerPositionX,
			PlayerPositionY: c.PlayerPositionY,
			BodySelection:   c.Payload.Selection,
		}

		var hitResponse nnsdk.HitEnemyResponse
		hitResponse, err = providers.HitEnemy(enemy, hitRequest)
		if err != nil {
			return err
		}

		// Verifico se il MOB è morto
		if hitResponse.EnemyDie == true {
			// TODO: ricevere monete e riconpensa

			// Aggiorno modifica del messaggio
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.mob_killed"),
			)

			var ok tgbotapi.InlineKeyboardMarkup
			ok = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						helpers.Trans(c.Player.Language.Slug, "continue"), "hunting.fight.return_map",
					),
				),
			)

			editMessage.ReplyMarkup = &ok

			// Setto stato
			c.Payload.Kill++
			c.Payload.InFight = false
			c.Payload.EnemyID = 0

			// Un mob è stato scofinto riaggiorno mappa e riaggiorno record su redis
			var maps nnsdk.Map
			maps, err = providers.GetMapByID(c.Payload.MapID)
			if err != nil {
				return err
			}

			// Registro mappa e posizione iniziale del player
			err = helpers.SetRedisMapHunting(maps)
			if err != nil {
				return err
			}

			// Invio messaggio
			_, err = services.SendMessage(editMessage)
			if err != nil {
				return err
			}

			return
		}

		// Verifico se il PLAYER è morto
		if hitResponse.PlayerDie == true {
			// Aggiorno messaggio notificando al player che è morto
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.player_killed"),
			)

			var ok tgbotapi.InlineKeyboardMarkup
			ok = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						helpers.Trans(c.Player.Language.Slug, "continue"), "hunting.fight.return_map",
					),
				),
			)

			editMessage.ReplyMarkup = &ok

			// Invio messaggio
			_, err = services.SendMessage(editMessage)
			if err != nil {
				return err
			}

			// Il player è morto non mi r
			c.State.Completed = helpers.SetTrue()

			// Invoco player Death
			// new(DeathController).Handle(c.Update)
			return
		}

		// Se ne il player e ne il mob è morto, continua lo scontro
		// Messagio di notifica per vedere risultato attacco
		if hitResponse.DodgeAttack == true {
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.miss", hitResponse.EnemyDamage),
			)
		} else {
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.damage", hitResponse.PlayerDamage, hitResponse.EnemyDamage),
			)
		}

		ok := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action"),
			),
		)
		editMessage.ReplyMarkup = &ok

	case "return_map":
		c.Payload.InFight = false

		// Trasformo la mappa in qualcosa di più leggibile su telegram
		var decodedMap string
		decodedMap, err = helpers.DecodeMapToDisplay(maps, c.PlayerPositionX, c.PlayerPositionY)
		if err != nil {
			return err
		}

		// Forzo invio messaggio contenente la mappa
		editMessage = services.NewEditMessage(
			c.Player.ChatID,
			c.Update.CallbackQuery.Message.MessageID,
			decodedMap,
		)

		editMessage.ParseMode = "HTML"
		editMessage.ReplyMarkup = &mapKeyboard
	case "no-action":
		//
	}

	// Non sono state fatte modifiche al messaggio
	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
		stats, _ := providers.GetPlayerStats(c.Player)
		editMessage = services.NewEditMessage(
			c.Player.ChatID,
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.Language.Slug, "combat.card",
				enemy.Name,
				enemy.LifePoint,
				enemy.LifeMax,
				c.Player.Username,
				*stats.LifePoint,
				100+stats.Level*10,
				helpers.Trans(c.Player.Language.Slug, bodyParts[c.Payload.Selection]),
			),
		)
		editMessage.ReplyMarkup = &mobKeyboard
	}

	// Invio messaggio modificato
	_, err = services.SendMessage(editMessage)

	return
}
