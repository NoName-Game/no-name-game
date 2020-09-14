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
	Handle(*pb.Player, tgbotapi.Update)
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
	CurrentState struct {
		Controller string
		Stage      int32
		Completed  bool
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
	c.LoadControllerData() // TODO: Verificare

	// Verifico se il player si trova in determinati stati non consentiti
	// e che quindi non permettano l'init del controller richiamato
	var inStateBlocked = c.InStatesBlocker() // TODO: Verificare
	if inStateBlocked {
		return false
	}

	// TODO: test
	if c.CurrentState.Controller, c.CurrentState.Stage, err = helpers.CheckStateNew(c.Player.ID, c.Configuration.Controller, c.CurrentState.Stage); err != nil {
		panic(err)
	}

	// Verifico lo stato della player
	// if c.PlayerData.CurrentState, _, err = helpers.CheckState(
	// 	c.Player,
	// 	c.PlayerData.ActiveStates,
	// 	c.Configuration.Controller,
	// 	c.Configuration.Payload,
	// 	c.ControllerFather,
	// ); err != nil {
	// 	panic(err)
	// }

	// Verifico se esistono condizioni per cambiare stato o uscire
	if c.BackTo(c.Configuration.ControllerBack.FromStage, c.Configuration.ControllerBack.To) {
		return false
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

// Validate - Metodo comune per mandare messaggio di validazione
func (c *BaseController) Validate() {
	// Se non ha un messaggio particolare allora setto configurazione di default
	if c.Validation.Message == "" {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	}

	// if c.Validation.ReplyKeyboard.Keyboard == nil {
	// 	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
	// 		tgbotapi.NewKeyboardButtonRow(
	// 			tgbotapi.NewKeyboardButton(
	// 				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
	// 			),
	// 		),
	// 	)
	// }

	// Invio il messaggio in caso di errore e chiudo
	validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
	validatorMsg.ParseMode = "markdown"
	if c.Validation.ReplyKeyboard.Keyboard != nil {
		validatorMsg.ReplyMarkup = c.Validation.ReplyKeyboard
	}

	_, err := services.SendMessage(validatorMsg)
	if err != nil {
		panic(err)
	}
}

// Breaking - Metodo che permette di verificare se si vogliono fare
// delle azioni che permetteranno di concludere
func (c *BaseController) BackTo(canBackFromStage int32, controller Controller) (backed bool) {
	// Verifico se è effetivamente un messaggio di testo e non una callback
	if c.Update.Message != nil {
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.back") {
			if c.Configuration.Controller != "" {
				if c.CurrentState.Stage > canBackFromStage {
					c.CurrentState.Stage = 0
					return
				}
			}

			// se è stato settato un controller esco
			if controller != nil {
				// Rimuovo testo messaggio
				c.Update.Message.Text = ""
				controller.Handle(c.Player, c.Update)
				backed = true
				return
			}

			// Cancello stato dalla memoria
			helpers.DelCacheState(c.Player.ID)
			helpers.DelCacheControllerStage(c.Player.ID, c.CurrentState.Controller)

			helpers.DelPayloadController(c.Player.ID, c.CurrentState.Controller)
			new(MenuController).Handle(c.Player, c.Update)
			backed = true
			return
		}

		// Abbandona - chiude definitivamente cancellando anche lo stato
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.clears") ||
			c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.more") {
			if !c.PlayerData.PlayerStats.GetDead() && c.Clearable() {
				// Cancello stato da cache
				helpers.DelCacheState(c.Player.ID)
				helpers.DelCacheControllerStage(c.Player.ID, c.CurrentState.Controller)

				helpers.DelPayloadController(c.Player.ID, c.CurrentState.Controller)

				// Cancello stato da ws
				if _, err := services.NnSDK.DeletePlayerStateByController(helpers.NewContext(1), &pb.DeletePlayerStateByControllerRequest{
					PlayerID:   c.Player.ID,
					Controller: c.CurrentState.Controller,
					Force:      true,
				}); err != nil {
					panic(err)
				}

				// Call menu controller
				new(MenuController).Handle(c.Player, c.Update)

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
			new(MenuController).Handle(c.Player, c.Update)

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
func (c *BaseController) Completing(payload interface{}) (err error) {
	// Controllo se posso aggiornare lo stato
	// TODO: da vedere forse si può togliere il controllo
	// Aggiorno cache state
	helpers.SetCacheControllerStage(c.Player.ID, c.CurrentState.Controller, c.CurrentState.Stage)

	// TODO: da verifice
	helpers.SetPayloadController(c.Player.ID, c.CurrentState.Controller, payload)

	// Verifico se lo stato è completato chiudo
	if c.CurrentState.Completed {
		// Cancello stato dalla memoria
		helpers.DelCacheState(c.Player.ID)
		helpers.DelCacheControllerStage(c.Player.ID, c.CurrentState.Controller)

		helpers.DelPayloadController(c.Player.ID, c.CurrentState.Controller)

		// Call menu controller
		if c.Configuration.ControllerBack.To != nil {
			c.Update.Message.Text = ""
			c.Configuration.ControllerBack.To.Handle(c.Player, c.Update)
			return
		}

		new(MenuController).Handle(c.Player, c.Update)
		return
	}

	// Verifico se si vuole forzare il menu
	if c.ForceBackTo {
		// Cancello stato dalla memoria
		helpers.DelCacheState(c.Player.ID)
		helpers.DelPayloadController(c.Player.ID, c.CurrentState.Controller)

		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update)
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

// InTutorial - Metodo che semplifica il controllo se il player si trova dentro un tutorial
func (c *BaseController) InTutorial() bool {
	for _, state := range c.PlayerData.ActiveStates {
		if !state.Completed {
			if c.PlayerData.CurrentState != nil {
				if state.Father == 0 || state.Father != c.PlayerData.CurrentState.GetFather() {
					if state.Controller == "route.tutorial" {
						return true
					}
				}
			}
		}
	}

	return false
}
