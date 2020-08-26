package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
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
		MapID             uint32
		EnemyID           uint32
		Selection         int32 // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		InFight           bool
		Kill              uint32
	}
	PlayerPositionX int32
	PlayerPositionY int32
	NeedUpdateState bool
}

// Settings generali
var (
	// Parti di corpo disponibili per l'attacco
	bodyParts = [4]string{"head", "chest", "gauntlets", "leg"}

	// Keyboard inline di esplorazione
	mapKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", "hunting.move.up")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", "hunting.move.left"),
			tgbotapi.NewInlineKeyboardButtonData("➡️", "hunting.move.right"),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬇️", "hunting.move.down")),
	)

	tresureKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬆️", "hunting.move.up")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", "hunting.move.left"),
			tgbotapi.NewInlineKeyboardButtonData("❓️", "hunting.move.action"),
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
			tgbotapi.NewInlineKeyboardButtonData("🗾", "hunting.fight.return_map"),
			tgbotapi.NewInlineKeyboardButtonData("⚔️", "hunting.fight.hit"),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", "hunting.fight.down")),
	)
)

// ====================================
// Handle
// ====================================
func (c *HuntingController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller:        "route.hunting",
		ControllerBlocked: []string{"mission"},
		Payload:           c.Payload,
		ControllerBack: ControllerBack{
			To:        &MenuController{},
			FromStage: 0,
		},
	}) {
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)

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

	// Forzo il controllo dell'update, questo set servirà
	// al controllo poco più sotto
	c.NeedUpdateState = true

	// Ok! Run!
	if !hasError {
		err = c.Stage()
		if err != nil {
			panic(err)
		}
	}

	// Verifico se c'è realmente bisogno di aggiornare lo stato,
	// effettuo questo controllo per evitare di fare un upadate inutile quando
	// mi muovo e basta
	if c.NeedUpdateState {
		// Aggiorno stato finale
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.PlayerData.CurrentState.Payload = string(payloadUpdated)

		var rUpdatePlayerState *pb.UpdatePlayerStateResponse
		rUpdatePlayerState, err = services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
			PlayerState: c.PlayerData.CurrentState,
		})
		if err != nil {
			panic(err)
		}
		c.PlayerData.CurrentState = rUpdatePlayerState.GetPlayerState()
	}

	// Verifico completamento aggiuntivo per cancellare il messaggio
	if c.PlayerData.CurrentState.GetCompleted() {
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
func (c *HuntingController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	// Il player deve avere sempre e perfoza un'arma equipaggiata
	// Indipendentemente dallo stato in cui si trovi
	if !helpers.CheckPlayerHaveOneEquippedWeapon(c.Player) {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "hunting.error.no_weapon_equipped")
		c.PlayerData.CurrentState.Completed = true
		return true, err
	}

	return false, err
}

// ====================================
// Stage Map -> Drop -> Finish
// ====================================
func (c *HuntingController) Stage() (err error) {
	switch c.PlayerData.CurrentState.Stage {
	// In questo stage faccio entrare il player nella mappa
	case 0:
		// Verifico se il player vuole uscire dalla caccia
		if c.Update.Message != nil {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "hunting.leave") {
				c.PlayerData.CurrentState.Completed = true
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
	if c.Payload.MapID == 0 || c.Update.Message != nil {
		// Se è qualsiasi messaggio diverso da hunting non lo calcolo
		// in quanto adnrebbe a generare più volte il messaggio con la stessa mappa
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.hunting") &&
			c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.tutorial") {
			return
		}

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
		// Dalla ultima posizione recupero il pianeta corrente
		var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
		rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		// Recupero dettagli della mappa e per non appesantire le chiamate
		// al DB registro il tutto sula cache
		var rGetMapByID *pb.GetMapByIDResponse
		rGetMapByID, err = services.NnSDK.GetMapByID(helpers.NewContext(1), &pb.GetMapByIDRequest{
			ID: rGetPlayerCurrentPlanet.GetPlanet().GetMapID(),
		})
		if err != nil {
			return err
		}

		var maps = rGetMapByID.GetMaps()

		// Registro mappa e posizione iniziale del player
		helpers.SetMapInCache(maps)
		helpers.SetCachedPlayerPositionInMap(maps, c.Player, "X", maps.GetStartPositionX())
		helpers.SetCachedPlayerPositionInMap(maps, c.Player, "Y", maps.GetStartPositionY())

		// Trasformo la mappa in qualcosa di più leggibile su telegram
		var decodedMap string
		decodedMap, err = helpers.DecodeMapToDisplay(maps, maps.GetStartPositionX(), maps.GetStartPositionY())
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

		return err
	}

	// Se il messaggio è di tipo callback ed esiste una mappa associato al payload
	// potrebbe essere un messaggio lanciato da tasiterino, quindi acconsento allo spostamento
	if c.Payload.MapID > 0 && c.Update.CallbackQuery != nil {
		var maps *pb.Maps
		maps, err = helpers.GetMapInCache(c.Payload.MapID)
		if err != nil {
			return err
		}

		// Recupero posizione player
		// var playerPositionX, playerPositionY int
		c.PlayerPositionX, err = helpers.GetCachedPlayerPositionInMap(maps, c.Player, "X")
		if err != nil {
			return err
		}

		c.PlayerPositionY, err = helpers.GetCachedPlayerPositionInMap(maps, c.Player, "Y")
		if err != nil {
			return err
		}

		// Controllo tipo di callback data - move / fight
		actionType := strings.Split(c.Update.CallbackQuery.Data, ".")

		// Verifica tipo di movimento e mi assicuro che non sia in combattimento
		if actionType[1] == "move" {
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
func (c *HuntingController) Move(action string, maps *pb.Maps) (err error) {
	// Refresh della mappa
	var cellGrid [][]bool
	err = json.Unmarshal([]byte(maps.CellGrid), &cellGrid)
	if err != nil {
		return err
	}

	// Eseguo azione
	switch action {
	case "up":
		if c.PlayerPositionX > 0 && !cellGrid[c.PlayerPositionX-1][c.PlayerPositionY] {
			c.PlayerPositionX--
			break
		}

		return
	case "down":
		// Se non è un muro posso proseguire
		if c.PlayerPositionX < int32(len(cellGrid)-1) && !cellGrid[c.PlayerPositionX+1][c.PlayerPositionY] {
			c.PlayerPositionX++
			break
		}

		return
	case "left":
		if c.PlayerPositionY > 0 && !cellGrid[c.PlayerPositionX][c.PlayerPositionY-1] {
			c.PlayerPositionY--
			break
		}

		return
	case "right":
		if c.PlayerPositionY < int32(len(cellGrid)-1) && !cellGrid[c.PlayerPositionX][c.PlayerPositionY+1] {
			c.PlayerPositionY++
			break
		}

		return
	case "action":
		// Verifico se si trova sopra un tesoro se così fosse lancio
		// chiamata per verificare il drop
		var nearTresure bool
		var tresure *pb.Tresure
		tresure, nearTresure = helpers.CheckForTresure(maps, c.PlayerPositionX, c.PlayerPositionY)
		if nearTresure {
			// random per definire se è un tesoro o una trappola :D
			// Chiamo WS e recupero tesoro
			var rDropTresure *pb.DropTresureResponse
			rDropTresure, err = services.NnSDK.DropTresure(helpers.NewContext(1), &pb.DropTresureRequest{
				TresureID: tresure.ID,
				PlayerID:  c.Player.ID,
			})
			if err != nil {
				return err
			}

			// Verifico cosa è tornato e rispondo
			var editMessage tgbotapi.EditMessageTextConfig
			if rDropTresure.GetResource().GetID() > 0 {
				editMessage = services.NewEditMessage(
					c.Player.ChatID,
					c.Update.CallbackQuery.Message.MessageID,
					helpers.Trans(c.Player.Language.Slug, "tresure.found.resource", rDropTresure.GetResource().GetName()),
				)
			} else if rDropTresure.GetItem().GetID() > 0 {
				itemFound := helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", rDropTresure.GetItem().GetSlug()))
				editMessage = services.NewEditMessage(
					c.Player.ChatID,
					c.Update.CallbackQuery.Message.MessageID,
					helpers.Trans(c.Player.Language.Slug, "tresure.found.item", itemFound),
				)
			} else if rDropTresure.GetTransaction().GetID() > 0 {
				editMessage = services.NewEditMessage(
					c.Player.ChatID,
					c.Update.CallbackQuery.Message.MessageID,
					helpers.Trans(c.Player.Language.Slug, "tresure.found.transaction", rDropTresure.GetTransaction().GetValue()),
				)
			} else if rDropTresure.GetTrap().GetID() > 0 {
				if rDropTresure.GetTrap().GetPlayerDie() {
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
				// Player sopravvive...
				editMessage = services.NewEditMessage(
					c.Player.ChatID,
					c.Update.CallbackQuery.Message.MessageID,
					helpers.Trans(c.Player.Language.Slug, "tresure.found.trap", rDropTresure.GetTrap().GetDamage()),
				)
			} else {
				// Non hai trovato nulla
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
			editMessage.ParseMode = "markdown"

			// Un tesoro è stato aperto, devo refreshare la mappa per cancellarlo
			err = c.RefreshMap()
			if err != nil {
				return err
			}

			_, err = services.SendMessage(editMessage)
			if err != nil {
				return err
			}

			return err
		}

		return err
	case "no-action":
		// No action
	default:
		err = errors.New("action not recognized")
		return err
	}

	// Aggiorno nuova posizione del player
	helpers.SetCachedPlayerPositionInMap(maps, c.Player, "X", c.PlayerPositionX)
	helpers.SetCachedPlayerPositionInMap(maps, c.Player, "Y", c.PlayerPositionY)

	// Trasformo la mappa in qualcosa di più leggibile su telegram
	var decodedMap string
	decodedMap, err = helpers.DecodeMapToDisplay(maps, c.PlayerPositionX, c.PlayerPositionY)
	if err != nil {
		return err
	}

	// Se l'azione è valida e completa aggiorno risultato
	msg := services.NewEditMessage(c.Player.ChatID, c.Update.CallbackQuery.Message.MessageID, decodedMap)

	// Se un player si trova sulla stessa posizione un mob o di un tesoro effettuo il controllo
	var nearMob, nearTresure bool
	_, nearMob = helpers.CheckForMob(maps, c.PlayerPositionX, c.PlayerPositionY)
	_, nearTresure = helpers.CheckForTresure(maps, c.PlayerPositionX, c.PlayerPositionY)
	if nearMob {
		msg.ReplyMarkup = &fightKeyboard
	} else if nearTresure {
		msg.ReplyMarkup = &tresureKeyboard
	} else {
		msg.ReplyMarkup = &mapKeyboard
	}

	msg.ParseMode = "HTML"
	_, err = services.SendMessage(msg)
	if err != nil {
		// Il bot crasha nel caso ci fossero bad request da parte di telegram,
		// penso sia opportuno solo in questo caso non pensare agli errori delle api che potrebbero causare crash non dettati da noi
		services.ErrorHandler("Hunting TGBOTAPI Error", err)
		return nil
	}

	// Visto che si è trattato solo di un movimento non è necessario aggiornare lo stato
	c.NeedUpdateState = false

	return
}

// ====================================
// Fight
// ====================================
func (c *HuntingController) Fight(action string, maps *pb.Maps) (err error) {
	var enemy *pb.Enemy
	var editMessage tgbotapi.EditMessageTextConfig

	if c.Payload.EnemyID > 0 {
		var rGetEnemyByID *pb.GetEnemyByIDResponse
		rGetEnemyByID, err = services.NnSDK.GetEnemyByID(helpers.NewContext(1), &pb.GetEnemyByIDRequest{
			ID: c.Payload.EnemyID,
		})
		if err != nil {
			return err
		}

		// Se impostato recupero informazioni più aggiornate del mob
		enemy = rGetEnemyByID.GetEnemy()
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
		var rHitEnemy *pb.HitEnemyResponse
		rHitEnemy, err = services.NnSDK.HitEnemy(helpers.NewContext(1), &pb.HitEnemyRequest{
			EnemyID:         enemy.GetID(),
			PlayerID:        c.Player.ID,
			PlayerPositionX: c.PlayerPositionX,
			PlayerPositionY: c.PlayerPositionY,
			BodySelection:   c.Payload.Selection,
		})
		if err != nil {
			return err
		}

		// Verifico se il MOB è morto
		if rHitEnemy.GetEnemyDie() {
			// Costruisco messaggio di recap del drop
			var dropRecap string

			if rHitEnemy.GetEnemyDrop().GetResource() != nil {
				dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.resource", rHitEnemy.GetEnemyDrop().GetResource().GetName())
			} else if rHitEnemy.GetEnemyDrop().GetItem() != nil {
				itemFound := helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", rHitEnemy.GetEnemyDrop().GetItem().GetSlug()))
				dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.item", itemFound)
			} else if rHitEnemy.GetEnemyDrop().GetTransaction() != nil {
				dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.transaction", rHitEnemy.GetEnemyDrop().GetTransaction().GetValue())
			} else {
				dropRecap += helpers.Trans(c.Player.Language.Slug, "combat.found.nothing")
			}

			// Aggiungo anche esperinza recuperata
			dropRecap += fmt.Sprintf("\n\n%s", helpers.Trans(c.Player.Language.Slug, "combat.experience", rHitEnemy.GetPlayerExperience()))

			// Aggiorno modifica del messaggio
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.mob_killed", enemy.Name, dropRecap),
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
			c.Payload.EnemyID = 0

			err = c.RefreshMap()
			if err != nil {
				return err
			}

			// Invio messaggio
			_, err = services.SendMessage(editMessage)
			if err != nil {
				return err
			}

			return err
		}

		// Verifico se il PLAYER è morto
		if rHitEnemy.GetPlayerDie() {
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
		if rHitEnemy.GetEnemyDodge() {
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.miss", rHitEnemy.GetEnemyDamage()),
			)
		} else if rHitEnemy.GetPlayerDodge() {
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.mob_miss", rHitEnemy.GetPlayerDamage()),
			)
		} else {
			editMessage = services.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.damage", rHitEnemy.GetPlayerDamage(), rHitEnemy.GetEnemyDamage()),
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
		c.Payload.EnemyID = 0

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
	case "player-die":
		// Il player è morto
		c.PlayerData.CurrentState.Completed = true

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
				enemy.Name, strings.ToUpper(enemy.Rarity.Slug),
				enemy.LifePoint,
				enemy.LifeMax,
				c.Player.Username,
				c.PlayerData.PlayerStats.GetLifePoint(),
				100+c.PlayerData.PlayerStats.GetLevel()*10,
				helpers.Trans(c.Player.Language.Slug, bodyParts[c.Payload.Selection]),
			),
		)
		editMessage.ParseMode = "markdown"
		editMessage.ReplyMarkup = &mobKeyboard
	}

	// Invio messaggio modificato
	_, err = services.SendMessage(editMessage)

	return
}

// RefreshMap - Necessario per refreshare la mappa in caso
// di sconfitta di mob o apertura di tesori.
func (c *HuntingController) RefreshMap() (err error) {
	// Un mob è stato scofinto riaggiorno mappa e riaggiorno record cache
	rGetMapByID, err := services.NnSDK.GetMapByID(helpers.NewContext(1), &pb.GetMapByIDRequest{
		ID: c.Payload.MapID,
	})
	if err != nil {
		return err
	}

	// Registro mappa e posizione iniziale del player
	helpers.SetMapInCache(rGetMapByID.GetMaps())
	return
}
