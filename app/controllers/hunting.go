package controllers

import (
	"encoding/json"
	"math/rand"
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
		Selection uint // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		InFight   bool
		Kill      uint
	}
	// Additional Data
	Callback *tgbotapi.CallbackQuery
}

// Settings generali
var (
	// Antiflood
	antiFloodSeconds float64 = 1.0

	// Parti di corpo disponibili per l'attacco
	bodyParts = [4]string{"head", "chest", "gauntlets", "leg"}

	// Keyboards
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
	// Verifico il tipo di messaggio
	if update.CallbackQuery != nil {
		c.Callback = update.CallbackQuery
	} else {
		c.Message = update.Message
	}

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
		validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
		validatorMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
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
	_, err = providers.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico se lo stato è completato chiudo
	if *c.State.Completed == true {
		_, err = providers.DeletePlayerState(c.State) // Delete
		if err != nil {
			panic(err)
		}

		err = helpers.DelRedisState(player)
		if err != nil {
			panic(err)
		}

		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update)
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

	// Al momento non ci sono particolari controlli da fare
	switch c.State.Stage {
	case 0:
		return false, err
	case 1:
		return false, err
	}

	return true, err
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

	// In questo stage notifico al player il completamento della mappa
	case 1:
		// Invio messaggio
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "hunting.complete"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.State.Completed = helpers.SetTrue()
	}

	return
}

// Hunting
func (c *HuntingController) Hunting() (err error) {
	// Se nel payload NON è presente un ID della mappa lo
	// recupero dalla posizione del player e invio al player il messaggio
	// principale contenente la mappa e il tastierino
	if c.Payload.MapID <= 0 {
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
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)

		return
	}

	// Recupero mappa da redis
	var maps nnsdk.Map
	maps, err = helpers.GetRedisMapHunting(c.Payload.MapID)
	if err != nil {
		return err
	}

	// Recupero posizione player
	var playerPositionX, playerPositionY int
	playerPositionX, err = helpers.GetRedisPlayerHuntingPosition(maps, c.Player, "X")
	playerPositionY, err = helpers.GetRedisPlayerHuntingPosition(maps, c.Player, "Y")

	// Blocker antiflood
	// if time.Since(maps.UpdatedAt).Seconds() > antiFloodSeconds {

	// Controllo tipo di callback data - move / fight
	actionType := strings.Split(c.Callback.Data, ".")

	// Verifica tipo di movimento e mi assicuro che non sia in combattimento
	if actionType[1] == "move" && !c.Payload.InFight {
		err = c.Move(actionType[2], maps, playerPositionX, playerPositionY)
	} else if actionType[1] == "fight" {
		err = c.Fight(actionType[2], maps, playerPositionX, playerPositionY)
	}

	if err != nil {
		return err
	}

	// Rimuove rotella di caricamento dal bottone
	services.AnswerCallbackQuery(services.NewAnswer(c.Callback.ID, "", false))
	return
	// }

	// Mostro errore antiflood
	// answer := services.NewAnswer(c.Callback.ID, "1 second delay", false)
	// services.AnswerCallbackQuery(answer)
	// return
}

// ====================================
// Movements
// ====================================
func (c *HuntingController) Move(action string, maps nnsdk.Map, playerPositionX int, playerPositionY int) (err error) {
	// Refresh della mappa
	var cellGrid [][]bool
	// var actionCompleted bool
	err = json.Unmarshal([]byte(maps.CellGrid), &cellGrid)
	if err != nil {
		return err
	}

	// Eseguo azione
	switch action {
	case "up":
		if !cellGrid[playerPositionX-1][playerPositionY] {
			playerPositionX--
			// actionCompleted = true
		} else {
			playerPositionX++
		}
	case "down":
		if !cellGrid[playerPositionX+1][playerPositionY] {
			playerPositionX++
			// actionCompleted = true
		} else {
			playerPositionX--
		}
	case "left":
		if !cellGrid[playerPositionX][playerPositionY-1] {
			playerPositionY--
			// actionCompleted = true
		} else {
			playerPositionY++
		}
	case "right":
		if !cellGrid[playerPositionX][playerPositionY+1] {
			playerPositionY++
			// actionCompleted = true
		} else {
			playerPositionY--
		}
	}

	// Aggiorno nuova posizione del player
	err = helpers.SetRedisPlayerHuntingPosition(maps, c.Player, "X", playerPositionX)
	err = helpers.SetRedisPlayerHuntingPosition(maps, c.Player, "Y", playerPositionY)
	if err != nil {
		return
	}

	// if actionCompleted {

	// Trasformo la mappa in qualcosa di più leggibile su telegram
	var decodedMap string
	decodedMap, err = helpers.DecodeMapToDisplay(maps, playerPositionX, playerPositionY)
	if err != nil {
		return err
	}

	// Se l'azione è valida e completa aggiorno risultato
	msg := services.NewEditMessage(c.Player.ChatID, c.Callback.Message.MessageID, decodedMap)

	// Se nella mappa viene visualizzato un mob allora mostro la fight keyboard
	// TODO: migliorare
	if strings.Contains(decodedMap, "*") {
		msg.ReplyMarkup = &fightKeyboard
	} else {
		msg.ReplyMarkup = &mapKeyboard
	}

	msg.ParseMode = "HTML"
	_, err = services.SendMessage(msg)
	if err != nil {
		return err
	}

	// }
	return
}

// ====================================
// Fight
// ====================================
func (c *HuntingController) Fight(action string, maps nnsdk.Map, playerPositionX int, playerPositionY int) (err error) {
	var enemy nnsdk.Enemy
	var editMessage tgbotapi.EditMessageTextConfig

	// Se impostato recupero informazioni più aggiornate del mob
	if c.Payload.EnemyID > 0 {
		enemy, err = providers.GetEnemyByID(c.Payload.EnemyID)
		if err != nil {
			return err
		}
	} else {
		// Recupero il mob più vicino con il quale combattere e me lo setto nel payload
		var coicheEnemyID int
		coicheEnemyID, err = helpers.ChooseEnemyInMap(maps, playerPositionX, playerPositionY)
		if err != nil {
			return err
		}
		enemy = maps.Enemies[coicheEnemyID]
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
		// DA RIVEDERE E SPOSTARE SUL WS
		mobDistance, _ := providers.Distance(maps, enemy)
		mobPercentage := (1000 - mobDistance) / 1000 // What percentage I see of the body? Number between 0 -> 1
		precision, _ := providers.PlayerPrecision(c.Player.ID, c.Payload.Selection)
		precision *= (85.0 / 37.0) * mobPercentage // Base precision

		// DA CHIEDERE
		if rand.Float64() < precision {
			// Hitted
			playerDamage, _ := providers.PlayerDamage(c.Player.ID)

			// DA SPOSTARE SUL WS
			damageToMob := uint(playerDamage)
			enemy.LifePoint = enemy.LifePoint - damageToMob

			// Mob ucciso
			if enemy.LifePoint > enemy.LifeMax || enemy.LifePoint == 0 {
				// Eseguo softdelete enemy
				_, err = providers.DeleteEnemy(enemy.ID)
				if err != nil {
					services.ErrorHandler("Cant delete enemy.", err)
				}

				// Aggiorno modifica del messaggio
				editMessage = services.NewEditMessage(c.Player.ChatID, c.Callback.Message.MessageID, helpers.Trans(c.Player.Language.Slug, "combat.mob_killed"))
				var ok tgbotapi.InlineKeyboardMarkup
				if c.Payload.Kill == uint(len(maps.Enemies)) {
					ok = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.finish")))
				} else {
					ok = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action")))
				}
				editMessage.ReplyMarkup = &ok

				// Incremento Statistiche player
				stats, _ := providers.GetPlayerStats(c.Player)
				helpers.IncrementExp(1, stats)

				// Setto stato
				c.Payload.Kill++
				c.Payload.InFight = false
			} else {
				// Il player subisce danno
				damageToPlayer, _ := providers.EnemyDamage(enemy.ID)
				stats, _ := providers.GetPlayerStats(c.Player)
				stats = helpers.DecrementLife(uint(damageToPlayer), stats)

				// Verifico se il danno ricevuto ha ucciso il player
				if *stats.LifePoint == 0 {
					c.State.Completed = helpers.SetTrue()

					// Invoco player Death
					// new(DeathController).Handle(c.Update)

					// TODO: devo impostare che il player è morto
					return
				} else {
					// Messagio di notifica per vedere risultato attacco
					editMessage = services.NewEditMessage(
						c.Player.ChatID,
						c.Callback.Message.MessageID,
						helpers.Trans(c.Player.Language.Slug, "combat.damage", damageToMob, uint(damageToPlayer)),
					)
					ok := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action"),
						),
					)
					editMessage.ReplyMarkup = &ok
				}

				// Aggiorno vita enemy
				_, err = providers.UpdateEnemy(enemy)
				if err != nil {
					services.ErrorHandler("Error while updating enemy", err)
				}

				// Aggiorno hunting map in redis in quanto contine informazione sui mob
				// helpers.UpdateHuntingMapRedis(maps, helpers.Player)
			}
		} else {
			// Schivata!
			damageToPlayer, _ := providers.EnemyDamage(enemy.ID)
			stats, _ := providers.GetPlayerStats(c.Player)
			stats = helpers.DecrementLife(uint(damageToPlayer), stats)

			if *stats.LifePoint == 0 {
				c.State.Completed = helpers.SetTrue()

				// Invoco player Death
				// new(DeathController).Handle(c.Update)
				return
			} else {
				editMessage = services.NewEditMessage(c.Player.ChatID, c.Callback.Message.MessageID, helpers.Trans(c.Player.Language.Slug, "combat.miss", damageToPlayer))
				ok := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action")))
				editMessage.ReplyMarkup = &ok
			}
		}
	case "returnMap":
		c.Payload.InFight = false

		// Trasformo la mappa in qualcosa di più leggibile su telegram
		var decodedMap string
		decodedMap, err = helpers.DecodeMapToDisplay(maps, playerPositionX, playerPositionY)
		if err != nil {
			return err
		}

		// Forzo invio messaggio contenente la mappa
		editMessage = services.NewEditMessage(
			c.Player.ChatID,
			c.Callback.Message.MessageID,
			decodedMap,
		)

		editMessage.ParseMode = "HTML"
		editMessage.ReplyMarkup = &mapKeyboard
	case "finish":
		c.State.Completed = helpers.SetTrue()

		services.SendMessage(services.NewEditMessage(c.Player.ChatID, c.Callback.Message.MessageID, helpers.Trans(c.Player.Language.Slug, "complete")))
		return
	case "no-action":
		//
	}

	// Non sono state fatte modifiche al messaggio
	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
		stats, _ := providers.GetPlayerStats(c.Player)
		editMessage = services.NewEditMessage(
			c.Player.ChatID,
			c.Callback.Message.MessageID,
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
	services.SendMessage(editMessage)

	// Aggiorno lo stato
	payloadUpdated, _ := json.Marshal(c.Payload)
	c.State.Payload = string(payloadUpdated)

	return
}
