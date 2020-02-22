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
		MapID     uint
		EnemyID   uint
		Selection int // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		InFight   bool
		Kill      uint
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
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗾", "hunting.fight.returnMap"),
			tgbotapi.NewInlineKeyboardButtonData("⚔", "hunting.fight.hit"),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", "hunting.fight.down")),
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
func (c *HuntingController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	// Il player deve avere sempre e perfoza un'arma equipaggiata
	// Indipendentemente dallo stato in cui si trovi
	if !helpers.CheckPlayerHaveOneEquippedWeapon(c.Player) {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "hunting.error.noWeaponEquipped")

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
		// if  c.Update.Message != nil {
		// 	if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.hunting") {
		// 		return
		// 	}
		// }

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
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno lo stato e ritorno
		c.Payload.MapID = maps.ID

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

	// Eseguo azione
	switch action {
	case "up":
		if !cellGrid[c.PlayerPositionX-1][c.PlayerPositionY] {
			c.PlayerPositionX--
		} else {
			c.PlayerPositionX++
		}
	case "down":
		if !cellGrid[c.PlayerPositionX+1][c.PlayerPositionY] {
			c.PlayerPositionX++
		} else {
			c.PlayerPositionX--
		}
	case "left":
		if !cellGrid[c.PlayerPositionX][c.PlayerPositionY-1] {
			c.PlayerPositionY--
		} else {
			c.PlayerPositionY++
		}
	case "right":
		if !cellGrid[c.PlayerPositionX][c.PlayerPositionY+1] {
			c.PlayerPositionY++
		} else {
			c.PlayerPositionY--
		}
	case "action":
		// Al momento viene usato per ucsire dalla mappa
		c.State.Completed = helpers.SetTrue()

		return
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

	// Se nella mappa viene visualizzato un mob allora mostro la fight keyboard
	// TODO: migliorare adesso verifica solo se ci sei sopra
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
						helpers.Trans(c.Player.Language.Slug, "continue"), "hunting.fight.returnMap",
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
						helpers.Trans(c.Player.Language.Slug, "continue"), "hunting.fight.returnMap",
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

	case "returnMap":
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
