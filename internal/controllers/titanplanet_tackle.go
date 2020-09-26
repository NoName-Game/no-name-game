package controllers

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"

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
		CallbackChatID    int64
		CallbackMessageID int
		TitanID           uint32
		Selection         int32 // 0: HEAD, 1: BODY, 2: ARMS, 3: LEGS
		InFight           bool
		InEvent           bool // Player have an event
		Kill              uint32
		EventID           uint32
	}
}

// Settings generali
var (
	titanKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔼", "titanplanet.tackle.fight.up")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚔️", "titanplanet.tackle.fight.hit"),
		),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔽", "titanplanet.tackle.fight.down")),
	)
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
		},
	}) {
		return
	}

	// Validate
	var hasError bool
	if hasError = c.Validator(); hasError {
		c.Validate()
		return
	}

	// Ok! Run!
	c.Stage()

	// Verifico completamento aggiuntivo per cancellare il messaggio
	if c.CurrentState.Completed {
		// Cancello messaggio contentente la mappa
		if err := helpers.DeleteMessage(c.Payload.CallbackChatID, c.Payload.CallbackMessageID); err != nil {
			c.Logger.Panic(err)
		}
	}

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
	// In questo stage faccio entrare il player nella mappa
	case 0:
		// Verifico se il player vuole uscire dalla caccia
		if c.Update.Message != nil {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.leave") {
				c.CurrentState.Completed = true
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
	// Recupero ultima posizione del player
	var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
	if rGetPlayerCurrentPlanet, err = config.App.Server.Connection.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Recupero titano in base alla posizione del player
	var rGetTitanByPlanetID *pb.GetTitanByPlanetIDResponse
	if rGetTitanByPlanetID, err = config.App.Server.Connection.GetTitanByPlanetID(helpers.NewContext(1), &pb.GetTitanByPlanetIDRequest{
		PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Se ricevo un messaggio normale probabilmente è un avvio o un abbandona
	if c.Update.Message != nil {
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

		// Preparo messaggio con la cardi di combattimento
		combactCard := helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.combat.card",
			rGetTitanByPlanetID.GetTitan().GetName(),
			rGetTitanByPlanetID.GetTitan().GetLifePoint(),
			rGetTitanByPlanetID.GetTitan().GetLifeMax(),
			c.Player.Username,
			c.Data.PlayerStats.GetLifePoint(),
			100+c.Data.PlayerStats.GetLevel()*10,
			helpers.Trans(c.Player.Language.Slug, bodyParts[c.Payload.Selection]),
		)

		// Invio quindi il mesaggio contenente le azioni disponibili
		msg := helpers.NewMessage(c.Player.ChatID, combactCard)
		msg.ReplyMarkup = titanKeyboard
		msg.ParseMode = "markdown"

		var tackleMessage tgbotapi.Message
		if tackleMessage, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno lo stato e ritorno
		c.Payload.CallbackChatID = tackleMessage.Chat.ID
		c.Payload.CallbackMessageID = tackleMessage.MessageID

		return
	}

	// Se il messaggio è di tipo callback sicuramete è un messaggio di attacco
	if c.Update.CallbackQuery != nil {
		// Verifico che non sia in corso un'evento
		if c.Payload.InEvent {
			// evento in corso
			var rGetEvent *pb.GetTitanEventByIDResponse
			if rGetEvent, err = config.App.Server.Connection.GetEventByID(helpers.NewContext(1), &pb.GetTitanEventByIDRequest{
				ID: c.Payload.EventID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			c.Event(c.Update.CallbackQuery.Data, rGetEvent.GetEvent(), rGetTitanByPlanetID.GetTitan())
		} else {
			// Controllo tipo di callback data - fight
			actionType := strings.Split(c.Update.CallbackQuery.Data, ".")

			// Verifica tipo di movimento e mi assicuro che non sia in combattimento
			if actionType[2] == "fight" {
				c.Fight(actionType[3], rGetTitanByPlanetID.GetTitan())
			}

			// Rimuove rotella di caricamento dal bottone
			err = helpers.AnswerCallbackQuery(
				helpers.NewAnswer(c.Update.CallbackQuery.ID, "", false),
			)

			return
		}
	}

	return
}

// ====================================
// Event
// ====================================
func (c *TitanPlanetTackleController) Event(text string, event *pb.TitanEvent, titan *pb.Titan) {
	var editMessage tgbotapi.EditMessageTextConfig
	// Standard message titanplanet.event.event1.choice1
	// route.event.eventID.choiceID
	actionType := strings.Split(c.Update.CallbackQuery.Data, ".")
	switch actionType[2] {
	case "fight":
		// arriverà dallo scontro, stampo semplicemente messaggio.
		editMessage = helpers.NewEditMessage(
			c.Player.GetChatID(),
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.GetLanguage().GetSlug(), event.TextCode),
		)
		var keyboardRow [][]tgbotapi.InlineKeyboardButton
		for _, choice := range event.Choices {
			keyboardRow = append(keyboardRow, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(helpers.Trans(c.Player.GetLanguage().GetSlug(), choice.GetTextCode()), choice.GetTextCode()),
			))
		}

		var ok = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: keyboardRow}
		editMessage.ReplyMarkup = &ok
	default:
		// Teoricamente è una choice
		if actionType[1] == "event" {
			// controllo che la choice faccia effettivamente parte dell'evento
			choiceID, err := strconv.Atoi(strings.Split(actionType[3], "choice")[1])
			if err != nil {
				c.Logger.Panic(err)
			}

			exist := false
			for _, choice := range event.Choices {
				if choice.ID == uint32(choiceID) {
					exist = true
				}
			}

			if exist {
				var rSubmitAnswer *pb.SubmitAnswerResponse
				if rSubmitAnswer, err = config.App.Server.Connection.SubmitAnswer(helpers.NewContext(1), &pb.SubmitAnswerRequest{
					TitanID:  titan.ID,
					ChoiceID: uint32(choiceID),
					PlayerID: c.Player.GetID(),
				}); err != nil {
					c.Logger.Panic(err)
				}

				if rSubmitAnswer.IsMalus {
					// Malus!
					// Player riceve danni
					editMessage = helpers.NewEditMessage(c.Player.ChatID, c.Update.CallbackQuery.Message.MessageID, helpers.Trans(c.Player.GetLanguage().GetSlug(), "titanplanet.event.wrong"))
					var ok = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData(
								helpers.Trans(c.Player.Language.Slug, "continue"), "titanplanet.tackle.fight.no_action",
							),
						),
					)
					editMessage.ReplyMarkup = &ok
					editMessage.ParseMode = tgbotapi.ModeMarkdown

				} else {
					// Bonus!
					// titano riceve danni
					if rSubmitAnswer.Hit.TitanDie {
						editMessage = helpers.NewEditMessage(
							c.Player.ChatID,
							c.Update.CallbackQuery.Message.MessageID,
							helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.combat.mob_killed", titan.GetName()),
						)

						var ok = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData(
									helpers.Trans(c.Player.Language.Slug, "continue"), "titanplanet.tackle.fight.titan_die",
								),
							),
						)
						editMessage.ParseMode = "markdown"
						editMessage.ReplyMarkup = &ok

						// Setto stato
						c.Payload.Kill++
						c.Payload.TitanID = 0
					} else {
						editMessage = helpers.NewEditMessage(
							c.Player.ChatID, c.Update.CallbackQuery.Message.MessageID, helpers.Trans(c.Player.Language.Slug, "titanplanet.event.correct", rSubmitAnswer.Hit.PlayerDamage))
						var ok = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData(
									helpers.Trans(c.Player.Language.Slug, "continue"), "titanplanet.tackle.fight.no_action",
								),
							),
						)
						editMessage.ReplyMarkup = &ok
						editMessage.ParseMode = tgbotapi.ModeMarkdown
					}
				}
				c.Payload.InEvent = false
			} else {
				// Risposta non presente fra quelle predefinite dall'evento. ERRORE
				c.Logger.Panic(errors.New("choice choosen not in event choices"))
			}
		}
	}
	// Non sono state fatte modifiche al messaggio
	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
		helpers.NewEditMessage(
			c.Player.GetChatID(),
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.GetLanguage().GetSlug(), event.TextCode),
		)
		editMessage.ParseMode = "markdown"
		editMessage.ReplyMarkup = &titanKeyboard
	}

	// Invio messaggio modificato
	if _, err := helpers.SendMessage(editMessage); err != nil {
		c.Logger.Panic(err)
	}

	return
}

// ====================================
// Fight
// ====================================
func (c *TitanPlanetTackleController) Fight(action string, titan *pb.Titan) {
	var err error
	var editMessage tgbotapi.EditMessageTextConfig

	switch action {
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
		if rHitTitan, err = config.App.Server.Connection.HitTitan(helpers.NewContext(1), &pb.HitTitanRequest{
			TitanID:       titan.GetID(),
			PlayerID:      c.Player.ID,
			BodySelection: c.Payload.Selection,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Verifico se il MOB è morto
		if rHitTitan.GetTitanDie() || titan.GetLifePoint() <= 0 {
			// Aggiorno modifica del messaggio
			editMessage = helpers.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.combat.mob_killed", titan.GetName()),
			)

			var ok = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						helpers.Trans(c.Player.Language.Slug, "continue"), "titanplanet.tackle.fight.titan_die",
					),
				),
			)
			editMessage.ParseMode = "markdown"
			editMessage.ReplyMarkup = &ok

			// Setto stato
			c.Payload.Kill++
			c.Payload.TitanID = 0

			// Invio messaggio
			if _, err = helpers.SendMessage(editMessage); err != nil {
				c.Logger.Panic(err)
			}

			return
		}

		// Verifico se il PLAYER è morto
		if rHitTitan.GetPlayerDie() {
			// Aggiorno messaggio notificando al player che è morto
			editMessage = helpers.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.player_killed"),
			)

			var ok = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						helpers.Trans(c.Player.Language.Slug, "continue"), "titanplanet.tackle.fight.player-die",
					),
				),
			)

			editMessage.ReplyMarkup = &ok

			// Invio messaggio
			if _, err = helpers.SendMessage(editMessage); err != nil {
				c.Logger.Panic(err)
			}

			return
		}

		// Se ne il player e ne il mob è morto, continua lo scontro
		// Messagio di notifica per vedere risultato attacco
		if rHitTitan.GetDodgeAttack() {
			editMessage = helpers.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.miss", rHitTitan.GetTitanDamage()),
			)
		} else {
			editMessage = helpers.NewEditMessage(
				c.Player.ChatID,
				c.Update.CallbackQuery.Message.MessageID,
				helpers.Trans(c.Player.Language.Slug, "combat.damage", rHitTitan.GetPlayerDamage(), rHitTitan.GetTitanDamage()),
			)
		}

		ok := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Ok!", "titanplanet.tackle.fight.no_action"),
			),
		)
		editMessage.ReplyMarkup = &ok
		// 15% probabilità che si scateni un evento al prossimo giro.
		r := rand.Int31n(101)
		if r <= 50 {
			c.Payload.InEvent = true
			// Recupero un evento random
			var rEventRandom *pb.GetRandomEventResponse
			if rEventRandom, err = config.App.Server.Connection.GetRandomEvent(helpers.NewContext(1), &pb.GetRandomEventRequest{}); err != nil {
				c.Logger.Panic(err)
			}
			c.Payload.EventID = rEventRandom.GetEvent().GetID()

		}
	case "player_die":
		// Il player è morto
		c.CurrentState.Completed = true
		return
	case "titan_die":
		// Il player è morto
		c.CurrentState.Completed = true
		// Drop Moment
		c.Drop(titan)
		return
	case "no_action":
		//

	}

	// Non sono state fatte modifiche al messaggio
	if editMessage == (tgbotapi.EditMessageTextConfig{}) {
		editMessage = helpers.NewEditMessage(
			c.Player.ChatID,
			c.Update.CallbackQuery.Message.MessageID,
			helpers.Trans(c.Player.Language.Slug, "titanplanet.tackle.combat.card",
				titan.GetName(),
				titan.GetLifePoint(),
				titan.GetLifeMax(),
				c.Player.Username,
				c.Data.PlayerStats.GetLifePoint(),
				100+c.Data.PlayerStats.GetLevel()*10,
				helpers.Trans(c.Player.Language.Slug, bodyParts[c.Payload.Selection]),
			),
		)
		editMessage.ParseMode = "markdown"
		editMessage.ReplyMarkup = &titanKeyboard
	}

	// Invio messaggio modificato
	if _, err = helpers.SendMessage(editMessage); err != nil {
		c.Logger.Panic(err)
	}

	return
}

func (c *TitanPlanetTackleController) Drop(titan *pb.Titan) {
	var err error
	// THIS FUNCTION TAKE ALL THE DAMAGES INFLICTED BY PLAYER AND GIVE HIM THE RIGHT DROP

	var rTitanDamage *pb.GetTitanDamageByTitanIDResponse
	if rTitanDamage, err = config.App.Server.Connection.GetTitanDamageByTitanID(helpers.NewContext(1), &pb.GetTitanDamageByTitanIDRequest{
		TitanID: titan.ID,
	}); err != nil {
		c.Logger.Panic(err)
	}
	for _, damage := range rTitanDamage.Damages {
		var rGetPlayer *pb.GetPlayerByIDResponse
		rGetPlayer, err = config.App.Server.Connection.GetPlayerByID(helpers.NewContext(1), &pb.GetPlayerByIDRequest{
			ID: damage.PlayerID,
		})
		if err != nil {
			c.Logger.Panic(err)
		}
		// Parte calcolo drop
		// TODO

		// Crafto messaggio drop
		msg := helpers.NewMessage(rGetPlayer.GetPlayer().ChatID, helpers.Trans(
			rGetPlayer.GetPlayer().GetLanguage().GetSlug(), "titanplanet.tackle.reward", damage.GetDamageInflicted() /*Aggiungere lista drop*/),
		)
		msg.ParseMode = tgbotapi.ModeMarkdown
		_, err := helpers.SendMessage(msg)
		if err != nil {
			c.Logger.Panic(err)
		}
	}

	return
}
