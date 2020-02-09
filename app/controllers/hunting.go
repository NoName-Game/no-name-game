package controllers

//
// import (
// 	"encoding/json"
// 	"math/rand"
// 	"strings"
// 	"time"
//
// 	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
//
// 	"bitbucket.org/no-name-game/nn-telegram/app/providers"
//
// 	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
// 	"bitbucket.org/no-name-game/nn-telegram/services"
// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
// )
//
// //====================================
// // HuntingController
// //====================================
// type HuntingController struct {
// 	BaseController
// 	Payload struct {
// 		IDMap     uint
// 		Selection uint // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
// 		IDEnemies uint
// 		InFight   bool
// 		Kill      uint
// 	}
// 	// Additional Data
// 	Callback   *tgbotapi.CallbackQuery
// }
//
// // Settings
// var (
// 	// Antiflood
// 	antiFloodSeconds float64 = 1.0
//
// 	// Parti di corpo disponibili per l'attacco
// 	bodyParts = [4]string{"head", "chest", "gauntlets", "leg"}
//
// 	// Keyboards
// 	mapKeyboard = tgbotapi.NewInlineKeyboardMarkup(
// 		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", "hunting.move.up")),
// 		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", "hunting.move.left"), tgbotapi.NewInlineKeyboardButtonData("⭕", "hunting.move.action"), tgbotapi.NewInlineKeyboardButtonData("➡️", "hunting.move.right")),
// 		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", "hunting.move.down")),
// 	)
// 	fightKeyboard = tgbotapi.NewInlineKeyboardMarkup(
// 		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", "hunting.move.up")),
// 		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", "hunting.move.left"), tgbotapi.NewInlineKeyboardButtonData("⚔️", "hunting.fight.start"), tgbotapi.NewInlineKeyboardButtonData("➡️", "hunting.move.right")),
// 		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", "hunting.move.down")),
// 	)
// 	mobKeyboard = tgbotapi.NewInlineKeyboardMarkup(
// 		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔼", "hunting.fight.up")),
// 		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🗾", "hunting.fight.returnMap"), tgbotapi.NewInlineKeyboardButtonData("⚔", "hunting.fight.hit")),
// 		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", "hunting.fight.down")),
// 	)
// )
//
// //====================================
// // Handle
// //====================================
// func (c *HuntingController) Handle(update tgbotapi.Update) {
// 	// Current Controller instance
// 	var err error
// 	c.RouteName, c.Update, c.Message = "route.hunting", update, update.Message
//
// 	// Check current state for this routes
// 	c.State, _ = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)
//
// 	// Set and load payload
// 	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)
//
// 	// Check message type
// 	if update.Message != nil {
// 		// Current Controller instance
// 		c.Message = update.Message
//
// 		// Go to validator
// 		if !c.Validator() {
// 			c.State, err = providers.UpdatePlayerState(c.State)
// 			if err != nil {
// 				services.ErrorHandler("Cant update player stats", err)
// 			}
//
// 			c.Stage()
// 			return
// 		}
//
// 		// Validator goes errors
// 		validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
// 		services.SendMessage(validatorMsg)
// 		return
// 	} else if update.CallbackQuery != nil {
// 		// Current Controller instance
// 		c.Callback = update.CallbackQuery
//
// 		c.Hunting()
// 		return
// 	}
//
// 	return
// }
//
// //====================================
// // Validator
// //====================================
// func (c *HuntingController) Validator() (hasErrors bool) {
// 	c.Validation.Message = helpers.Trans("validationMessage")
//
// 	// Il player deve avere sempre e perfoza un'arma equipaggiata
// 	// Indipendentemente dallo stato
// 	if !helpers.CheckPlayerHaveOneEquippedWeapon(helpers.Player) {
// 		c.Validation.Message = helpers.Trans("hunting.error.noWeaponEquipped")
//
// 		//====================================
// 		// FORCED COMPLETE!
// 		//====================================
// 		helpers.FinishAndCompleteState(c.State, helpers.Player)
// 		//====================================
//
// 		return true
// 	}
//
// 	switch c.State.Stage {
// 	case 0:
// 		return false
// 	case 1:
// 		return false
// 	}
//
// 	return true
// }
//
// //====================================
// // Stage Waiting -> Map -> Drop -> Finish
// //====================================
// func (c *HuntingController) Stage() {
// 	switch c.State.Stage {
// 	case 0:
// 		// Join Map
// 		c.Hunting()
// 	case 1:
// 		// Invio messaggio
// 		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("hunting.complete"))
// 		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
// 			tgbotapi.NewKeyboardButtonRow(
// 				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
// 			),
// 		)
// 		services.SendMessage(msg)
//
// 		//====================================
// 		// COMPLETE!
// 		//====================================
// 		helpers.FinishAndCompleteState(c.State, helpers.Player)
// 		//====================================
// 	}
// }
//
// func (c *HuntingController) Hunting() {
// 	var err error
//
// 	// Recupero mappa da redis, se non esiste l'istanza la creo
// 	huntingMap, isNew := helpers.GetHuntingMapRedis(c.Payload.IDMap, helpers.Player)
// 	if isNew {
// 		// Invio messaggio contenente la mappa
// 		msg := services.NewMessage(helpers.Player.ChatID, helpers.TextDisplay(huntingMap))
// 		msg.ReplyMarkup = mapKeyboard
// 		msg.ParseMode = "HTML"
// 		services.SendMessage(msg)
//
// 		// Aggiorno stato
// 		c.Payload.IDMap = huntingMap.ID
// 		payloadUpdated, _ := json.Marshal(c.Payload)
// 		c.State.Payload = string(payloadUpdated)
// 		c.State, err = providers.UpdatePlayerState(c.State)
// 		if err != nil {
// 			services.ErrorHandler("Cant update player stats", err)
// 		}
// 		return
// 	}
//
// 	// Blocker antiflood
// 	if time.Since(huntingMap.UpdatedAt).Seconds() > antiFloodSeconds {
// 		// Controllo tipo di callback data - move / fight
// 		actionType := strings.Split(c.Callback.Data, ".")
//
// 		// Verifica tipo di movimento e mi assicuro che non sia in combattimento
// 		if actionType[1] == "move" && !c.Payload.InFight {
// 			c.move(actionType[2], huntingMap)
// 		} else if actionType[1] == "fight" {
// 			c.fight(actionType[2], huntingMap)
// 		}
//
// 		// Rimuove rotella di caricamento dal bottone
// 		services.AnswerCallbackQuery(services.NewAnswer(c.Callback.ID, "", false))
// 		return
// 	}
//
// 	// Mostro errore antiflood
// 	answer := services.NewAnswer(c.Callback.ID, "1 second delay", false)
// 	services.AnswerCallbackQuery(answer)
// 	return
// }
//
// //====================================
// // Movements
// //====================================
// func (c *HuntingController) move(action string, huntingMap nnsdk.Map) {
// 	// Refresh della mappa
// 	var cellMap [66][66]bool
// 	var actionCompleted bool
// 	err := json.Unmarshal([]byte(huntingMap.Cell), &cellMap)
// 	if err != nil {
// 		services.ErrorHandler("Error unmarshal map", err)
// 	}
//
// 	// Eseguo azione
// 	switch action {
// 	case "up":
// 		if !cellMap[huntingMap.PlayerX-1][huntingMap.PlayerY] {
// 			huntingMap.PlayerX--
// 			actionCompleted = true
// 		} else {
// 			huntingMap.PlayerX++
// 		}
// 	case "down":
// 		if !cellMap[huntingMap.PlayerX+1][huntingMap.PlayerY] {
// 			huntingMap.PlayerX++
// 			actionCompleted = true
// 		} else {
// 			huntingMap.PlayerX--
// 		}
// 	case "left":
// 		if !cellMap[huntingMap.PlayerX][huntingMap.PlayerY-1] {
// 			huntingMap.PlayerY--
// 			actionCompleted = true
// 		} else {
// 			huntingMap.PlayerY++
// 		}
// 	case "right":
// 		if !cellMap[huntingMap.PlayerX][huntingMap.PlayerY+1] {
// 			huntingMap.PlayerY++
// 			actionCompleted = true
// 		} else {
// 			huntingMap.PlayerY--
// 		}
// 	}
//
// 	// Aggiorno orario (serve per il controllo del delay) e aggiorno record su redis
// 	huntingMap.UpdatedAt = time.Now()
//
// 	// Se l'azione è valida e completa aggiorno risultato
// 	if actionCompleted {
// 		msg := services.NewEditMessage(helpers.Player.ChatID, c.Callback.Message.MessageID, helpers.TextDisplay(huntingMap))
// 		if strings.Contains(helpers.TextDisplay(huntingMap), "*") {
// 			msg.ReplyMarkup = &fightKeyboard
// 		} else {
// 			msg.ReplyMarkup = &mapKeyboard
// 		}
//
// 		msg.ParseMode = "HTML"
// 		services.SendMessage(msg)
// 	}
//
// 	// Aggiorno redis map
// 	helpers.UpdateHuntingMapRedis(huntingMap, helpers.Player)
// }
//
// //====================================
// // Fight
// //====================================
// func (c *HuntingController) fight(action string, huntingMap nnsdk.Map) {
// 	var err error
// 	var enemy nnsdk.Enemy
// 	var editMessage tgbotapi.EditMessageTextConfig
//
// 	// Se impostato recupero informazioni più aggiornate del mob
// 	if c.Payload.IDEnemies > 0 {
// 		enemy, _ = providers.GetEnemyByID(c.Payload.IDEnemies)
// 	} else {
// 		// Recupero il mob più vicino con il quale combattere e me lo setto nel payload
// 		enemy = huntingMap.Enemies[helpers.ChooseMob(huntingMap)]
// 	}
//
// 	switch action {
// 	// Avvio di un nuovo combattimento
// 	case "start":
// 		// Setto nuove informazioni stato
// 		c.Payload.IDEnemies = enemy.ID
// 		c.Payload.InFight = true
// 	case "up":
// 		// Setto nuova parte del corpo da colpire
// 		if c.Payload.Selection > 0 {
// 			c.Payload.Selection--
// 		} else {
// 			c.Payload.Selection = 3
// 		}
// 	case "down":
// 		// Setto nuova parte del corpo da colpire
// 		if c.Payload.Selection < 3 {
// 			c.Payload.Selection++
// 		} else {
// 			c.Payload.Selection = 0
// 		}
// 	case "hit":
// 		// DA RIVEDERE E SPOSTARE SUL WS
// 		mobDistance, _ := providers.Distance(huntingMap, enemy)
// 		mobPercentage := (1000 - mobDistance) / 1000 // What percentage I see of the body? Number between 0 -> 1
// 		precision, _ := providers.PlayerPrecision(helpers.Player.ID, c.Payload.Selection)
// 		precision *= (85.0 / 37.0) * mobPercentage // Base precision
//
// 		// DA CHIEDERE
// 		if rand.Float64() < precision {
// 			// Hitted
// 			playerDamage, _ := providers.PlayerDamage(helpers.Player.ID)
//
// 			// DA SPOSTARE SUL WS
// 			damageToMob := uint(playerDamage)
// 			enemy.LifePoint = enemy.LifePoint - damageToMob
//
// 			// Mob ucciso
// 			if enemy.LifePoint > enemy.LifeMax || enemy.LifePoint == 0 {
// 				// Eseguo softdelete enemy
// 				_, err = providers.DeleteEnemy(enemy.ID)
// 				if err != nil {
// 					services.ErrorHandler("Cant delete enemy.", err)
// 				}
//
// 				// Aggiorno modifica del messaggio
// 				editMessage = services.NewEditMessage(helpers.Player.ChatID, c.Callback.Message.MessageID, helpers.Trans("combat.mob_killed"))
// 				var ok tgbotapi.InlineKeyboardMarkup
// 				if c.Payload.Kill == uint(len(huntingMap.Enemies)) {
// 					ok = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.finish")))
// 				} else {
// 					ok = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action")))
// 				}
// 				editMessage.ReplyMarkup = &ok
//
// 				// Incremento Statistiche player
// 				stats, _ := providers.GetPlayerStats(helpers.Player)
// 				helpers.IncrementExp(1, stats)
//
// 				// Setto stato
// 				c.Payload.Kill++
// 				c.Payload.InFight = false
// 			} else {
// 				// Il player subisce danno
// 				damageToPlayer, _ := providers.EnemyDamage(enemy.ID)
// 				stats, _ := providers.GetPlayerStats(helpers.Player)
// 				stats = helpers.DecrementLife(uint(damageToPlayer), stats)
//
// 				// Verifico se il danno ricevuto ha ucciso il player
// 				if *stats.LifePoint == 0 {
// 					//====================================
// 					// COMPLETE! - Il player non è più in grado di fare nulla, esco!
// 					//====================================
// 					helpers.FinishAndCompleteState(c.State, helpers.Player)
// 					//====================================
//
// 					// Invoco player Death
// 					new(DeathController).Handle(c.Update)
// 					return
// 				} else {
// 					// Messagio di notifica per vedere risultato attacco
// 					editMessage = services.NewEditMessage(
// 						helpers.Player.ChatID,
// 						c.Callback.Message.MessageID,
// 						helpers.Trans("combat.damage", damageToMob, uint(damageToPlayer)),
// 					)
// 					ok := tgbotapi.NewInlineKeyboardMarkup(
// 						tgbotapi.NewInlineKeyboardRow(
// 							tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action"),
// 						),
// 					)
// 					editMessage.ReplyMarkup = &ok
// 				}
//
// 				// Aggiorno vita enemy
// 				_, err = providers.UpdateEnemy(enemy)
// 				if err != nil {
// 					services.ErrorHandler("Error while updating enemy", err)
// 				}
//
// 				// Aggiorno hunting map in redis in quanto contine informazione sui mob
// 				helpers.UpdateHuntingMapRedis(huntingMap, helpers.Player)
// 			}
// 		} else {
// 			// Schivata!
// 			damageToPlayer, _ := providers.EnemyDamage(enemy.ID)
// 			stats, _ := providers.GetPlayerStats(helpers.Player)
// 			stats = helpers.DecrementLife(uint(damageToPlayer), stats)
//
// 			if *stats.LifePoint == 0 {
// 				//====================================
// 				// COMPLETE! - Il player non è più in grado di fare nulla, esco!
// 				//====================================
// 				helpers.FinishAndCompleteState(c.State, helpers.Player)
// 				//====================================
//
// 				// Invoco player Death
// 				new(DeathController).Handle(c.Update)
// 				return
// 			} else {
// 				editMessage = services.NewEditMessage(helpers.Player.ChatID, c.Callback.Message.MessageID, helpers.Trans("combat.miss", damageToPlayer))
// 				ok := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ok!", "hunting.fight.no-action")))
// 				editMessage.ReplyMarkup = &ok
// 			}
// 		}
// 	case "returnMap":
// 		c.Payload.InFight = false
//
// 		// Forzo invio messaggio contenente la mappa
// 		editMessage = services.NewEditMessage(
// 			helpers.Player.ChatID,
// 			c.Callback.Message.MessageID,
// 			helpers.TextDisplay(huntingMap),
// 		)
// 		editMessage.ParseMode = "HTML"
// 		editMessage.ReplyMarkup = &mapKeyboard
// 	case "finish":
// 		//====================================
// 		// COMPLETE!
// 		//====================================
// 		helpers.FinishAndCompleteState(c.State, helpers.Player)
// 		//====================================
//
// 		services.SendMessage(services.NewEditMessage(helpers.Player.ChatID, c.Callback.Message.MessageID, helpers.Trans("complete")))
// 		return
// 	case "no-action":
// 		//
// 	}
//
// 	// Non sono state fatte modifiche al messaggio
// 	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
// 		stats, _ := providers.GetPlayerStats(helpers.Player)
// 		editMessage = services.NewEditMessage(
// 			helpers.Player.ChatID,
// 			c.Callback.Message.MessageID,
// 			helpers.Trans("combat.card",
// 				enemy.Name,
// 				enemy.LifePoint,
// 				enemy.LifeMax,
// 				helpers.Player.Username,
// 				*stats.LifePoint,
// 				(100+stats.Level*10),
// 				helpers.Trans(bodyParts[c.Payload.Selection]),
// 			),
// 		)
// 		editMessage.ReplyMarkup = &mobKeyboard
// 	}
//
// 	// Invio messaggio modificato
// 	services.SendMessage(editMessage)
//
// 	// Aggiorno lo stato
// 	payloadUpdated, _ := json.Marshal(c.Payload)
// 	c.State.Payload = string(payloadUpdated)
// 	c.State, err = providers.UpdatePlayerState(c.State)
// 	if err != nil {
// 		services.ErrorHandler("Cant update player stats", err)
// 	}
// }
