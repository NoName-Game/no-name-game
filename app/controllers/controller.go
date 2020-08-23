package controllers

import (
	"fmt"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/services"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	// UnClearables - Lista delle rotte che non devono subire gli effetti di abbandona anche se forzati a mano
	UnClearables = []string{"route.hunting"}
)

// Controller - Intereffaccia base di tutti i controller
type Controller interface {
	Handle(*pb.Player, tgbotapi.Update, bool)
	Validator()
	Stage()
}

type BaseController struct {
	Player     *pb.Player
	Update     tgbotapi.Update
	Validation struct {
		HasErrors     bool
		Message       string
		ReplyKeyboard tgbotapi.ReplyKeyboardMarkup
	}
	PlayerData struct {
		ActiveStates []*pb.PlayerState
		PlayerStats  *pb.PlayerStats
		CurrentState *pb.PlayerState
	}
	ControllerFather uint32
	ForceBackTo      bool
	Configuration    ControllerConfiguration
}

type ControllerConfiguration struct {
	ControllerBlocked []string
	ControllerBack    ControllerBack
	Controller        string
	Payload           interface{}
	ProxyStatment     bool // Viene richiamato da qualche altro controller
}

type ControllerBack struct {
	To        Controller
	FromStage int32
}

func (c *BaseController) InitController(configuration ControllerConfiguration) bool {
	var err error

	// Associo configurazione che tutti i controller dovrebbero avere
	c.Configuration = configuration

	// Carico controller data
	c.LoadControllerData()

	// Verifico se il player si trova in determinati stati non consentiti
	// e che quindi non permettano l'init del controller richiamato
	var inStateBlocked = c.InStatesBlocker()
	if inStateBlocked {
		return false
	}

	// Verifico lo stato della player
	if c.PlayerData.CurrentState, _, err = helpers.CheckState(
		c.Player,
		c.PlayerData.ActiveStates,
		c.Configuration.Controller,
		c.Configuration.Payload,
		c.ControllerFather,
	); err != nil {
		panic(err)
	}

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !c.Configuration.ProxyStatment {
		if c.BackTo(c.Configuration.ControllerBack.FromStage, c.Configuration.ControllerBack.To) {
			return false
		}
	}

	return true
}

// Carico controller data
func (c *BaseController) LoadControllerData() {
	// Recupero stato utente
	rGetActivePlayerStates, err := services.NnSDK.GetActivePlayerStates(helpers.NewContext(1), &pb.GetActivePlayerStatesRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}
	c.PlayerData.ActiveStates = rGetActivePlayerStates.GetStates()

	// Recupero stats utente
	rGetPlayerStats, err := services.NnSDK.GetPlayerStats(helpers.NewContext(1), &pb.GetPlayerStatsRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}
	c.PlayerData.PlayerStats = rGetPlayerStats.GetPlayerStats()
}

// Breaking - Metodo che permette di verificare se si vogliono fare
// delle azioni che permetteranno di concludere
func (c *BaseController) BackTo(canBackFromStage int32, controller Controller) (backed bool) {
	// Verifico se è effetivamente un messaggio di testo e non una callback
	if c.Update.Message != nil {
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.back") {
			if c.Configuration.Controller != "" {
				if c.PlayerData.CurrentState.GetStage() <= canBackFromStage {
					// Cancello stato da cache
					helpers.DelCacheState(c.Player.ID)

					// Cancello record a db
					if c.PlayerData.CurrentState != nil {
						_, err := services.NnSDK.DeletePlayerState(helpers.NewContext(1), &pb.DeletePlayerStateRequest{
							PlayerStateID: c.PlayerData.CurrentState.ID,
							ForceDelete:   true,
						})

						if err != nil {
							panic(err)
						}
					}

					controller.Handle(c.Player, c.Update, true)
					backed = true
					return
				}

				c.PlayerData.CurrentState.Stage = 0
				return
			}

			// Cancello stato da cache
			helpers.DelCacheState(c.Player.ID)

			// Cancello record a db
			_, err := services.NnSDK.DeletePlayerState(helpers.NewContext(1), &pb.DeletePlayerStateRequest{
				PlayerStateID: c.PlayerData.CurrentState.ID,
				ForceDelete:   true,
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
			if !c.PlayerData.PlayerStats.GetDead() && c.Clearable() {
				// Cancello stato da cache
				helpers.DelCacheState(c.Player.ID)

				// Cancello record a db
				if c.PlayerData.CurrentState != nil {
					_, err := services.NnSDK.DeletePlayerState(helpers.NewContext(1), &pb.DeletePlayerStateRequest{
						PlayerStateID: c.PlayerData.CurrentState.ID,
						ForceDelete:   true,
					})
					if err != nil {
						panic(err)
					}
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
			// Cancello stato dalla memoria
			helpers.DelCacheState(c.Player.ID)

			// Call menu controller
			new(MenuController).Handle(c.Player, c.Update, true)

			backed = true
			return
		}
	}

	return
}

// Clearable
func (c *BaseController) Clearable() (clearable bool) {
	// Certi controller non devono subire la cancellazione degli stati
	// perchè magari hanno logiche particolari o lo gestiscono a loro modo
	for _, state := range c.PlayerData.ActiveStates {
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
	if c.PlayerData.CurrentState.GetCompleted() {
		// Posso cancellare lo stato solo se non è figlio di qualche altro stato
		if c.PlayerData.CurrentState.GetFather() == 0 {
			_, err = services.NnSDK.DeletePlayerState(helpers.NewContext(1), &pb.DeletePlayerStateRequest{
				PlayerStateID: c.PlayerData.CurrentState.ID,
				ForceDelete:   true,
			})
			if err != nil {
				return err
			}
		}

		// Cancello stato dalla memoria
		helpers.DelCacheState(c.Player.ID)

		// Call menu controller
		if c.Configuration.ControllerBack.To != nil {
			c.Configuration.ControllerBack.To.Handle(c.Player, c.Update, true)
			return
		}

		new(MenuController).Handle(c.Player, c.Update, true)
		return
	}

	// Verifico se si vuole forzare il menu
	if c.ForceBackTo {
		// Cancello stato dalla memoria
		helpers.DelCacheState(c.Player.ID)

		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update, true)
	}

	return
}

// InStatesBlocker
// Certi controller non possono essere eseguiti se il player si trova in determinati stati.
// Ogni controller ha la possibilità nell'handle di passare la lista di rotte bloccanti per esso.
func (c *BaseController) InStatesBlocker() (inStates bool) {
	// Certi controller non devono subire la cancellazione degli stati
	// perchè magari hanno logiche particolari o lo gestiscono a loro modo
	for _, state := range c.PlayerData.ActiveStates {
		// Verifico se non fa parte dello stesso padre e che lo stato non sia completato
		if !state.Completed {
			if c.PlayerData.CurrentState != nil {
				if state.Father == 0 || state.Father != c.PlayerData.CurrentState.GetFather() {
					for _, blockState := range c.Configuration.ControllerBlocked {
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
