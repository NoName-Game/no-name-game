package controllers

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type HuntingController struct {
	Update     tgbotapi.Update
	Message    *tgbotapi.Message
	Callback   *tgbotapi.CallbackQuery
	RouteName  string
	Validation struct {
		HasErrors bool
		Message   string
	}
	Payload struct {
		IDMap     uint
		Selection uint // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		IDEnemies uint
		InFight   bool
		Kill      uint
	}
	// Additional Data
	State nnsdk.PlayerState
}

var (
	// Settings
	antiFloodSeconds float64 = 1.0

	// Keyboards
	mapKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", "hunting.move.up")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", "hunting.move.left"), tgbotapi.NewInlineKeyboardButtonData("⭕", "hunting.move.action"), tgbotapi.NewInlineKeyboardButtonData("➡️", "hunting.move.right")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", "hunting.move.down")),
	)
	fightKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", "hunting.move.up")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", "hunting.move.left"), tgbotapi.NewInlineKeyboardButtonData("⚔️", "hunting.fight.start"), tgbotapi.NewInlineKeyboardButtonData("➡️", "hunting.move.right")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", "hunting.move.down")),
	)
	mobKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔼", "hunting.fight.up")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🗾", "hunting.fight.returnMap"), tgbotapi.NewInlineKeyboardButtonData("⚔", "hunting.fight.hit")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", "hunting.fight.down")),
	)
)

//====================================
// Handle
//====================================
func (c *HuntingController) Handle(update tgbotapi.Update) {
	c.RouteName = "route.hunting"
	c.Update = update
	c.State, _ = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// Check message type
	if update.Message != nil {
		// Current Controller instance
		c.Message = update.Message

		// Go to validator
		c.Validation.HasErrors, c.State = c.Validator(c.State)
		if !c.Validation.HasErrors {
			c.State, _ = providers.UpdatePlayerState(c.State)
			c.Stage(c.State)
			return
		}

		// Validator goes errors
		validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
		services.SendMessage(validatorMsg)
		return
	} else if update.CallbackQuery != nil {
		// Current Controller instance
		c.Callback = update.CallbackQuery

		c.Hunting()
		return
	}

	return
}

//====================================
// Validator
//====================================
func (c *HuntingController) Validator(state nnsdk.PlayerState) (hasErrors bool, newState nnsdk.PlayerState) {
	c.Validation.Message = helpers.Trans("validationMessage")

	// Il player deve avere sempre e perfoza un'arma equipaggiata
	// Indipendentemente dallo stato
	if !helpers.CheckPlayerHaveOneEquippedWeapon(helpers.Player) {
		c.Validation.Message = helpers.Trans("hunting.error.noWeaponEquipped")
		helpers.FinishAndCompleteState(state, helpers.Player)
		return true, state
	}

	switch state.Stage {
	case 0:
		return false, state
	case 1:
		return false, state
	}

	return true, state
}

//====================================
// Stage Waiting -> Map -> Drop -> Finish
//====================================
func (c *HuntingController) Stage(state nnsdk.PlayerState) {
	switch state.Stage {
	case 0:
		// Join Map
		c.Hunting()
	case 1:
		//====================================
		// IMPORTANT!
		//====================================
		helpers.FinishAndCompleteState(state, helpers.Player)
		//====================================

		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("hunting.complete"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)
	}
}

func (c *HuntingController) Hunting() {
	// Recupero mappa da redis, se non esiste l'istanza la creo
	huntingMap, isNew := helpers.GetHuntingMapRedis(c.Payload.IDMap, helpers.Player)
	if isNew {
		c.Payload.IDMap = huntingMap.ID
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, _ = providers.UpdatePlayerState(c.State)

		msg := services.NewMessage(helpers.Player.ChatID, helpers.TextDisplay(huntingMap))
		msg.ReplyMarkup = mapKeyboard
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
		return
	}

	// Blocker antiflood
	if time.Since(huntingMap.UpdatedAt).Seconds() > antiFloodSeconds {
		// Controllo tipo di callback data - move / fight
		actionType := strings.Split(c.Callback.Data, ".")

		// Verifica tipo di movimento e mi assicuro che non sia in combattimento
		if actionType[1] == "move" && !c.Payload.InFight {
			c.move(actionType[2], huntingMap)
		} else if actionType[1] == "fight" {
			c.fight(actionType[2], huntingMap)
		}

		// Rimuove rotella di caricamento dal bottone
		services.AnswerCallbackQuery(services.NewAnswer(c.Callback.ID, "", false))
		return
	}

	// Mostro errore antiflood
	answer := services.NewAnswer(c.Callback.ID, "1 second delay", false)
	services.AnswerCallbackQuery(answer)
	return
}

//====================================
// Movements
//====================================
func (c *HuntingController) move(action string, huntingMap nnsdk.Map) {
	// Refresh della mappa
	var cellMap [66][66]bool
	var actionCompleted bool
	err := json.Unmarshal([]byte(huntingMap.Cell), &cellMap)
	if err != nil {
		services.ErrorHandler("Error during unmarshal", err)
	}

	// Eseguo azione
	switch action {
	case "up":
		if !cellMap[huntingMap.PlayerX-1][huntingMap.PlayerY] {
			huntingMap.PlayerX--
			actionCompleted = true
		} else {
			huntingMap.PlayerX++
		}
	case "down":
		if !cellMap[huntingMap.PlayerX+1][huntingMap.PlayerY] {
			huntingMap.PlayerX++
			actionCompleted = true
		} else {
			huntingMap.PlayerX--
		}
	case "left":
		if !cellMap[huntingMap.PlayerX][huntingMap.PlayerY-1] {
			huntingMap.PlayerY--
			actionCompleted = true
		} else {
			huntingMap.PlayerY++
		}
	case "right":
		if !cellMap[huntingMap.PlayerX][huntingMap.PlayerY+1] {
			huntingMap.PlayerY++
			actionCompleted = true
		} else {
			huntingMap.PlayerY--
		}
	}

	// Aggiorno orario (serve per il controllo del delay) e aggiorno record su redis
	huntingMap.UpdatedAt = time.Now()
	helpers.UpdateHuntingMapRedis(huntingMap, helpers.Player)

	if actionCompleted {
		msg := services.NewEditMessage(helpers.Player.ChatID, c.Callback.Message.MessageID, helpers.TextDisplay(huntingMap))
		if strings.Contains(helpers.TextDisplay(huntingMap), "*") {
			msg.ReplyMarkup = &fightKeyboard
		} else {
			msg.ReplyMarkup = &mapKeyboard
		}

		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	}
}

//====================================
// Fight
//====================================
func (c *HuntingController) fight(action string, huntingMap nnsdk.Map) {
	bodyParts := [4]string{"head", "chest", "gauntlets", "leg"}

	var editMessage tgbotapi.EditMessageTextConfig
	var enemy nnsdk.Enemy

	// Se impostato recupero informazioni più aggiornate del mob
	if c.Payload.IDEnemies > 0 {
		enemy, _ = providers.GetEnemyByID(c.Payload.IDEnemies)
	} else {
		// Recupero il mob più vicino con il quale combattere e me lo setto nel payload
		enemy = huntingMap.Enemies[helpers.ChooseMob(huntingMap)]
	}

	switch action {
	case "start":
		// Aggiorno lo stato
		c.Payload.IDEnemies = enemy.ID
		c.Payload.InFight = true
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, _ = providers.UpdatePlayerState(c.State)
	case "up":
		if c.Payload.Selection > 0 {
			c.Payload.Selection--
		} else {
			c.Payload.Selection = 3
		}
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, _ = providers.UpdatePlayerState(c.State)
	case "down":
		if c.Payload.Selection < 3 {
			c.Payload.Selection++
		} else {
			c.Payload.Selection = 0
		}
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, _ = providers.UpdatePlayerState(c.State)
	case "hit":

		// DA RIVEDERE E SPOSTARE SUL WS
		mobDistance, _ := providers.Distance(huntingMap, enemy)
		mobPercentage := ((1000 - mobDistance) / 1000) // What percentage I see of the body? Number between 0 -> 1
		precision, _ := providers.PlayerPrecision(helpers.Player.ID, c.Payload.Selection)
		precision *= (85.0 / 37.0) * mobPercentage // Base precision

		// DA CHIEDERE
		if rand.Float64() < precision {
			// Hitted
			// ARMI GIÀ RECUPERATE
			_, err := providers.GetPlayerWeapons(helpers.Player, "true")
			if err != nil {
				services.ErrorHandler("Error while retriving weapons", err)
			}
			playerDamage, _ := providers.PlayerDamage(helpers.Player.ID)

			// DA SPOSTARE SUL WS
			damageToMob := uint(playerDamage)
			enemy.LifePoint = enemy.LifePoint - damageToMob

			if enemy.LifePoint > enemy.LifeMax || enemy.LifePoint == 0 {
				// enemy die
				c.Payload.Kill++
				enemy.LifePoint = 0
				_, err = providers.DeleteEnemy(enemy.ID)
				if err != nil {
					services.ErrorHandler("Cant delete enemy.", err)
				}
				editMessage = services.NewEditMessage(helpers.Player.ChatID, c.Callback.Message.MessageID, helpers.Trans("combat.mob_killed"))
				var ok tgbotapi.InlineKeyboardMarkup
				if c.Payload.Kill == uint(len(huntingMap.Enemies)) {
					ok = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.finish")))
				} else {
					ok = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action")))
				}
				editMessage.ReplyMarkup = &ok
				// Add drop
				stats, _ := providers.GetPlayerStats(helpers.Player)
				helpers.IncrementExp(1, stats)

				c.Payload.InFight = false
				payloadUpdated, _ := json.Marshal(c.Payload)
				c.State.Payload = string(payloadUpdated)
				c.State, _ = providers.UpdatePlayerState(c.State)
			} else {
				damageToPlayer, _ := providers.EnemyDamage(enemy.ID)
				stats, _ := providers.GetPlayerStats(helpers.Player)
				stats = helpers.DecrementLife(uint(damageToPlayer), stats)
				if *stats.LifePoint == 0 {
					// Player Die
					helpers.DeleteRedisAndDbState(helpers.Player)
					PlayerDeath(c.Update)
				} else {
					// Messagio di notifica per vedere risultato attacco
					editMessage = services.NewEditMessage(
						helpers.Player.ChatID,
						c.Callback.Message.MessageID,
						helpers.Trans("combat.damage", damageToMob, uint(damageToPlayer)),
					)
					ok := tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action"),
						),
					)
					editMessage.ReplyMarkup = &ok
				}

				_, err = providers.UpdateEnemy(enemy)
				if err != nil {
					services.ErrorHandler("Error while updating enemy", err)
				}

				helpers.UpdateHuntingMapRedis(huntingMap, helpers.Player)
			}
		} else {
			// Miss by player
			damageToPlayer, _ := providers.EnemyDamage(enemy.ID)
			stats, _ := providers.GetPlayerStats(helpers.Player)
			stats = helpers.DecrementLife(uint(damageToPlayer), stats)
			if *stats.LifePoint == 0 {
				// Player Die
				helpers.DeleteRedisAndDbState(helpers.Player)
				PlayerDeath(c.Update)
			} else {
				editMessage = services.NewEditMessage(helpers.Player.ChatID, c.Callback.Message.MessageID, helpers.Trans("combat.miss", damageToPlayer))
				ok := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action")))
				editMessage.ReplyMarkup = &ok
			}
		}
	case "returnMap":
		c.Payload.InFight = false
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, _ = providers.UpdatePlayerState(c.State)

		// Forzo invio messaggio contenente la mappa
		editMessage = services.NewEditMessage(
			helpers.Player.ChatID,
			c.Callback.Message.MessageID,
			helpers.TextDisplay(huntingMap),
		)
		editMessage.ParseMode = "HTML"
		editMessage.ReplyMarkup = &mapKeyboard
	case "finish":
		helpers.FinishAndCompleteState(c.State, helpers.Player)
		services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, c.Callback.Message.MessageID, helpers.Trans("complete")))
		return
	case "no-action":
		//
	}

	// Standard Message
	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
		stats, _ := providers.GetPlayerStats(helpers.Player)
		editMessage = services.NewEditMessage(
			helpers.Player.ChatID,
			c.Callback.Message.MessageID,
			helpers.Trans("combat.card",
				enemy.Name,
				enemy.LifePoint,
				enemy.LifeMax,
				helpers.Player.Username,
				*stats.LifePoint,
				(100+stats.Level*10),
				helpers.Trans(bodyParts[c.Payload.Selection]),
			),
		)
		editMessage.ReplyMarkup = &mobKeyboard
	}

	services.SendMessage(editMessage)
}
