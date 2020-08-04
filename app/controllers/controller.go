package controllers

import (
	"fmt"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/services"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	// Lista delle rotte che non devono subire
	// gli effetti di abbandona anche se forzati a mano
	UnClearables = []string{"route.hunting"}
)

// Controller - Intereffaccia base di tutti i controller
type Controller interface {
	Handle(*pb.Player, tgbotapi.Update, bool)
	Validator()
	Stage()
}

type BaseController struct {
	Update     tgbotapi.Update
	Controller string
	Father     uint32
	Validation struct {
		HasErrors     bool
		Message       string
		ReplyKeyboard tgbotapi.ReplyKeyboardMarkup
	}
	Player       *pb.Player
	PlayerStats  *pb.PlayerStats
	CurrentState *pb.PlayerState
	ActiveStates []*pb.PlayerState
	Breaker      struct {
		ToMenu bool
	}
	ProxyStatment bool
}

func (c *BaseController) InitController(controller string, payload interface{}, blockers []string, player *pb.Player, update tgbotapi.Update) (initialized bool) {
	var err error
	initialized = true

	// Inizializzo variabili del controller
	c.Controller, c.Player, c.Update = controller, player, update

	// Recupero stato utente
	rGetActivePlayerStates, err := services.NnSDK.GetActivePlayerStates(helpers.NewContext(1), &pb.GetActivePlayerStatesRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}
	c.ActiveStates = rGetActivePlayerStates.GetStates()

	// Recupero stats utente per verificare se è morto
	rGetPlayerStats, err := services.NnSDK.GetPlayerStats(helpers.NewContext(1), &pb.GetPlayerStatsRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}
	c.PlayerStats = rGetPlayerStats.GetPlayerStats()

	// Verifico se il player si trova in determinati stati non consentiti
	// e che quindi non permettano l'init del controller richiamato
	var inStateBlocked = c.InStatesBlocker(blockers)
	if inStateBlocked {
		initialized = false
		return
	}

	// Verifico lo stato della player
	c.CurrentState, _, err = helpers.CheckState(*player, c.ActiveStates, c.Controller, payload, c.Father)

	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	return
}

// Breaking - Metodo che permette di verificare se si vogliono fare
// delle azioni che permetteranno di concludere
func (c *BaseController) BackTo(canBackFrom int32, controller Controller) (backed bool) {
	if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.back") {
		if c.Controller != "" {
			if c.CurrentState.GetStage() <= canBackFrom {
				// Cancello stato da redis
				_ = helpers.DelRedisState(*c.Player)

				// Cancello record a db
				_, err := services.NnSDK.DeletePlayerState(helpers.NewContext(1), &pb.DeletePlayerStateRequest{
					PlayerState: c.CurrentState,
				})
				if err != nil {
					panic(err)
				}

				controller.Handle(c.Player, c.Update, true)
				backed = true
				return
			}

			c.CurrentState.Stage = 0
			return
		}

		// Cancello stato da redis
		_ = helpers.DelRedisState(*c.Player)

		// Cancello record a db
		_, err := services.NnSDK.DeletePlayerState(helpers.NewContext(1), &pb.DeletePlayerStateRequest{
			PlayerState: c.CurrentState,
		})
		if err != nil {
			panic(err)
		}

		controller.Handle(c.Player, c.Update, true)
		backed = true
		return
	}

	// Abbandona - chiude definitivamente cancellando anche lo stato
	if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.clears") ||
		c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.more") {
		if !c.PlayerStats.GetDead() && c.Clearable() {
			// Cancello stato da redis
			_ = helpers.DelRedisState(*c.Player)

			// Cancello record a db
			_, err := services.NnSDK.DeletePlayerState(helpers.NewContext(1), &pb.DeletePlayerStateRequest{
				PlayerState: c.CurrentState,
			})
			if err != nil {
				panic(err)
			}

			// Call menu controller
			new(MenuController).Handle(c.Player, c.Update, true)

			backed = true
			return
		}
	}

	// Continua - mantiene lo stato attivo ma ti forza a tornare al menù
	// usato principalemente per notificare che esiste già un'attività in corso (Es. Missione)
	if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.continue") {
		// Cancello stato da redis
		_ = helpers.DelRedisState(*c.Player)

		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update, true)

		backed = true
		return
	}

	return
}

// Clearable
func (c *BaseController) Clearable() (clearable bool) {
	// Certi controller non devono subire la cancellazione degli stati
	// perchè magari hanno logiche particolari o lo gestiscono a loro modo
	for _, state := range c.ActiveStates {
		for _, unclearable := range UnClearables {
			if helpers.Trans(c.Player.Language.Slug, state.Controller) == helpers.Trans(c.Player.Language.Slug, unclearable) {
				return false

			}
		}
	}

	return true
}

// Completing - Metodo per settare il completamento di uno stato
func (c *BaseController) Completing() (err error) {
	// Verifico se lo stato è completato chiudo
	if c.CurrentState.GetCompleted() {
		// Posso cancellare lo stato solo se non è figlio di qualche altro stato
		if c.CurrentState.GetFather() == 0 {
			_, err := services.NnSDK.DeletePlayerState(helpers.NewContext(1), &pb.DeletePlayerStateRequest{
				PlayerState: c.CurrentState,
			})
			if err != nil {
				return err
			}
		}

		// Cancello stato da redis
		err = helpers.DelRedisState(*c.Player)
		if err != nil {
			panic(err)
		}

		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update, true)

		return
	}

	// Verifico se si vuole forzare il menu
	if c.Breaker.ToMenu {
		// Cancello stato da redis
		err = helpers.DelRedisState(*c.Player)
		if err != nil {
			panic(err)
		}

		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update, true)
	}

	return
}

// InStatesBlocker
// Certi controller non possono essere eseguiti se il player si trova in determinati stati.
// Ogni controller ha la possibilità nell'handle di passare la lista di rotte bloccanti per esso.
func (c *BaseController) InStatesBlocker(blockStates []string) (inStates bool) {
	// Certi controller non devono subire la cancellazione degli stati
	// perchè magari hanno logiche particolari o lo gestiscono a loro modo
	for _, state := range c.ActiveStates {
		// Verifico se non fa parte dello stesso padre e che lo stato non sia completato
		if !state.Completed {
			if c.CurrentState != nil {
				if state.Father == 0 || state.Father != c.CurrentState.GetFather() {
					for _, blockState := range blockStates {
						if helpers.Trans(c.Player.Language.Slug, state.Controller) == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("route.%s", blockState)) {
							msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "valodator.controller.blocked"))
							_, err := services.SendMessage(msg)
							if err != nil {
								panic(err)
							}

							return true
						}
					}
				}
			}

		}
	}

	return false
}
