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
// ====================================
// In questo controller il player avrà la possibilità di esplorare
// la mappa del pianeta che sta visitando, e di conseguenza affrontare mob,
// recupeare tesori e cascare in delle trappole
// ====================================

type HuntingController struct {
	BaseController
	Payload struct {
		CallbackMessageID int
		MapID             uint32
		PlayerPositionX   int32
		PlayerPositionY   int32
		BodySelection     int32
	}
}

// ====================================
// HuntingController - Settings
// ====================================
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
		ControllerBack: ControllerBack{
			To:        &MenuController{},
			FromStage: 0,
		},
	}, &c.Payload) {
		return
	}

	// Ok! Run!
	if err = c.Stage(); err != nil {
		panic(err)
	}

	// Verifico completamento aggiuntivo per cancellare il messaggio
	if c.CurrentState.Completed {
		// Cancello messaggio contentente la mappa
		if err = services.DeleteMessage(c.Player.ChatID, c.Payload.CallbackMessageID); err != nil {
			panic(err)
		}
	}

	// Completo progressione
	if err = c.Completing(&c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *HuntingController) Validator() (hasErrors bool) {
	return false
}

// ====================================
// Stage Map -> Drop -> Finish
// ====================================
func (c *HuntingController) Stage() (err error) {
	switch c.CurrentState.Stage {
	// In questo stage faccio entrare il player nella mappa
	case 0:
		// Verifico se il player vuole uscire dalla caccia
		if c.Update.Message != nil {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "hunting.leave") {
				c.CurrentState.Completed = true
				return err
			}
		}

		// Avvio ufficialmente la caccia!
		if err = c.Hunting(); err != nil {
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
	if c.Update.CallbackQuery == nil && c.Update.Message != nil {
		// Se è qualsiasi messaggio diverso da hunting non lo calcolo
		// in quanto adnrebbe a generare più volte il messaggio con la stessa mappa
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.hunting") &&
			c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.tutorial") &&
			c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.tutorial.continue") {
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
		if _, err = services.SendMessage(initHunting); err != nil {
			return err
		}

		// Recupero ultima posizione del player, dando per scontato che sia
		// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
		// Dalla ultima posizione recupero il pianeta corrente
		var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
		if rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			return err
		}

		// Recupero dettagli della mappa e per non appesantire le chiamate
		// al DB registro il tutto sula cache
		var rGetMapByID *pb.GetMapByIDResponse
		if rGetMapByID, err = services.NnSDK.GetMapByID(helpers.NewContext(1), &pb.GetMapByIDRequest{
			MapID: rGetPlayerCurrentPlanet.GetPlanet().GetMapID(),
		}); err != nil {
			return err
		}

		var maps = rGetMapByID.GetMaps()

		// Registro mappa e posizione iniziale del player
		helpers.SetMapInCache(maps)
		c.Payload.PlayerPositionX = maps.GetStartPositionX()
		c.Payload.PlayerPositionY = maps.GetStartPositionY()
		// helpers.SetCachedPlayerPositionInMap(maps, c.Player, "X", maps.GetStartPositionX())
		// helpers.SetCachedPlayerPositionInMap(maps, c.Player, "Y", maps.GetStartPositionY())

		// Trasformo la mappa in qualcosa di più leggibile su telegram
		var decodedMap string
		if decodedMap, err = helpers.DecodeMapToDisplay(maps, maps.GetStartPositionX(), maps.GetStartPositionY()); err != nil {
			return err
		}

		// Invio quindi il mesaggio contenente mappa e azioni disponibili
		msg := services.NewMessage(c.Player.ChatID, decodedMap)
		msg.ReplyMarkup = mapKeyboard
		msg.ParseMode = "HTML"

		var huntingMessage tgbotapi.Message
		if huntingMessage, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorno lo stato e ritorno
		c.Payload.CallbackMessageID = huntingMessage.MessageID
		c.Payload.MapID = maps.ID
		// helpers.SetHuntingCacheData(c.Player.ID, &HuntingCacheData{
		// 	CallbackMessageID: huntingMessage.MessageID,
		// 	MapID:             maps.ID,
		// })

		return err
	}

	// Se il messaggio è di tipo callback ed esiste una mappa associato al payload
	// potrebbe essere un messaggio lanciato da tasiterino, quindi acconsento allo spostamento
	if c.Update.CallbackQuery != nil && c.Update.Message == nil {
		var maps *pb.Maps
		if maps, err = helpers.GetMapInCache(c.Payload.MapID); err != nil {
			return err
		}

		// Recupero posizione player
		// var playerPositionX, playerPositionY int
		// if c.PlayerPositionX, err = helpers.GetCachedPlayerPositionInMap(maps, c.Player, "X"); err != nil {
		// 	return err
		// }
		//
		// if c.PlayerPositionY, err = helpers.GetCachedPlayerPositionInMap(maps, c.Player, "Y"); err != nil {
		// 	return err
		// }

		// Controllo tipo di callback data - move / fight
		actionType := strings.Split(c.Update.CallbackQuery.Data, ".")

		// Verifica tipo di movimento e mi assicuro che non sia in combattimento
		if actionType[1] == "move" {
			err = c.movements(actionType[2], maps)
		} else if actionType[1] == "fight" {
			err = c.fight(actionType[2], maps)
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
func (c *HuntingController) movements(action string, maps *pb.Maps) (err error) {
	// Refresh della mappa
	var cellGrid [][]bool
	if err = json.Unmarshal([]byte(maps.CellGrid), &cellGrid); err != nil {
		return
	}

	// Eseguo azione
	switch action {
	case "up":
		if c.Payload.PlayerPositionX > 0 && !cellGrid[c.Payload.PlayerPositionX-1][c.Payload.PlayerPositionY] {
			c.Payload.PlayerPositionX--
			break
		}

		return
	case "down":
		// Se non è un muro posso proseguire
		if c.Payload.PlayerPositionX < int32(len(cellGrid)-1) && !cellGrid[c.Payload.PlayerPositionX+1][c.Payload.PlayerPositionY] {
			c.Payload.PlayerPositionX++
			break
		}

		return
	case "left":
		if c.Payload.PlayerPositionY > 0 && !cellGrid[c.Payload.PlayerPositionX][c.Payload.PlayerPositionY-1] {
			c.Payload.PlayerPositionY--
			break
		}

		return
	case "right":
		if c.Payload.PlayerPositionY < int32(len(cellGrid)-1) && !cellGrid[c.Payload.PlayerPositionX][c.Payload.PlayerPositionY+1] {
			c.Payload.PlayerPositionY++
			break
		}

		return
	case "action":
		// Verifico se si trova sopra un tesoro se così fosse lancio
		// chiamata per verificare il drop
		var nearTresure bool
		var tresure *pb.Tresure
		tresure, nearTresure = helpers.CheckForTresure(maps, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY)
		if nearTresure {
			// random per definire se è un tesoro o una trappola :D
			// Chiamo WS e recupero tesoro
			var rDropTresure *pb.DropTresureResponse
			if rDropTresure, err = services.NnSDK.DropTresure(helpers.NewContext(1), &pb.DropTresureRequest{
				TresureID: tresure.ID,
				PlayerID:  c.Player.ID,
			}); err != nil {
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
					if _, err = services.SendMessage(editMessage); err != nil {
						return
					}

					return
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
			if err = c.RefreshMap(maps.ID); err != nil {
				return
			}

			if _, err = services.SendMessage(editMessage); err != nil {
				return
			}

			return
		}

		return
	case "no-action":
		// No action
	default:
		return errors.New("action not recognized")
	}

	// Aggiorno nuova posizione del player
	// helpers.SetCachedPlayerPositionInMap(maps, c.Player, "X", c.PlayerPositionX)
	// helpers.SetCachedPlayerPositionInMap(maps, c.Player, "Y", c.PlayerPositionY)

	// Trasformo la mappa in qualcosa di più leggibile su telegram
	var decodedMap string
	if decodedMap, err = helpers.DecodeMapToDisplay(maps, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY); err != nil {
		return
	}

	// Se l'azione è valida e completa aggiorno risultato
	msg := services.NewEditMessage(c.Player.ChatID, c.Update.CallbackQuery.Message.MessageID, decodedMap)

	// Se un player si trova sulla stessa posizione un mob o di un tesoro effettuo il controllo
	var nearMob, nearTresure bool
	_, nearMob = helpers.CheckForMob(maps, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY)
	_, nearTresure = helpers.CheckForTresure(maps, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY)
	if nearMob {
		msg.ReplyMarkup = &fightKeyboard
	} else if nearTresure {
		msg.ReplyMarkup = &tresureKeyboard
	} else {
		msg.ReplyMarkup = &mapKeyboard
	}

	msg.ParseMode = "HTML"
	if _, err = services.SendMessage(msg); err != nil {
		// Il bot crasha nel caso ci fossero bad request da parte di telegram,
		// penso sia opportuno solo in questo caso non pensare agli errori delle api che potrebbero causare crash non dettati da noi
		services.ErrorHandler("Hunting TGBOTAPI Error", err)
		return nil
	}

	return
}

// ====================================
// Fight
// ====================================
func (c *HuntingController) fight(action string, maps *pb.Maps) (err error) {
	var enemy *pb.Enemy
	var editMessage tgbotapi.EditMessageTextConfig

	// Recupero dettagli aggiornati enemy
	enemy, _ = helpers.CheckForMob(maps, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY)
	if enemy != nil {
		var rGetEnemyByID *pb.GetEnemyByIDResponse
		if rGetEnemyByID, err = services.NnSDK.GetEnemyByID(helpers.NewContext(1), &pb.GetEnemyByIDRequest{
			EnemyID: enemy.ID,
		}); err != nil {
			return err
		}
		enemy = rGetEnemyByID.GetEnemy()
	}

	switch action {
	// Avvio di un nuovo combattimento
	case "start":
		// In questo metodo non c'è niente da fare procedo con il stampare la card del combattimento
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
		// Effettuo chiamata al ws e recupero response dell'attacco
		var rHitEnemy *pb.HitEnemyResponse
		rHitEnemy, err = services.NnSDK.HitEnemy(helpers.NewContext(1), &pb.HitEnemyRequest{
			EnemyID:       enemy.GetID(),
			PlayerID:      c.Player.ID,
			BodySelection: c.Payload.BodySelection,
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

			// Se il mob è morto è necessario aggiornare la mappa
			if err = c.RefreshMap(maps.ID); err != nil {
				return
			}

			// Invio messaggio
			if _, err = services.SendMessage(editMessage); err != nil {
				return
			}

			return
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
			if _, err = services.SendMessage(editMessage); err != nil {
				return
			}

			return
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
		// Trasformo la mappa in qualcosa di più leggibile su telegram
		var decodedMap string
		decodedMap, err = helpers.DecodeMapToDisplay(maps, c.Payload.PlayerPositionX, c.Payload.PlayerPositionY)
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
		c.CurrentState.Completed = true

		return
	case "no-action":
		//
	}

	// Edit message viene passato vuoto solo se non si tratta di hit o bodyselection
	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
		editMessage = services.NewEditMessage(
			c.Player.ChatID,
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.Language.Slug, "combat.card",
				enemy.Name, strings.ToUpper(enemy.Rarity.Slug),
				enemy.LifePoint,
				enemy.LifeMax,
				c.Player.Username,
				c.Data.PlayerStats.GetLifePoint(),
				100+c.Data.PlayerStats.GetLevel()*10,
				helpers.Trans(c.Player.Language.Slug, bodyParts[c.Payload.BodySelection]),
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
func (c *HuntingController) RefreshMap(MapID uint32) (err error) {
	// Un mob è stato scofinto riaggiorno mappa e riaggiorno record cache
	var rGetMapByID *pb.GetMapByIDResponse
	if rGetMapByID, err = services.NnSDK.GetMapByID(helpers.NewContext(1), &pb.GetMapByIDRequest{
		MapID: MapID,
	}); err != nil {
		return
	}

	// Registro mappa e posizione iniziale del player
	helpers.SetMapInCache(rGetMapByID.GetMaps())
	return
}
