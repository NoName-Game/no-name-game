package controllers

import (
	"fmt"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/services"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Controller - Intereffaccia base di tutti i controller
type ControllerInterface interface {
	Handle(*pb.Player, tgbotapi.Update)
	Validator()
	Stage()
}

type Controller struct {
	Player     *pb.Player
	Update     tgbotapi.Update
	Validation struct {
		HasErrors     bool
		Message       string
		ReplyKeyboard tgbotapi.ReplyKeyboardMarkup
	}
	CurrentState ControllerCurrentState
	Data         struct {
		PlayerActiveStates []*pb.PlayerState
		PlayerStats        *pb.PlayerStats
	}
	ForceBackTo    bool
	Configurations ControllerConfigurations
}

type ControllerCurrentState struct {
	Controller string
	Stage      int32
	Completed  bool
	Payload    interface{}
}

type ControllerConfigurations struct {
	ControllerBlocked []string
	ControllerBack    ControllerBack
}

type ControllerBack struct {
	To        ControllerInterface
	FromStage int32
}

func (c *Controller) InitController(controller Controller) bool {
	var err error

	// Carico configurazione controller
	*c = controller

	// Carico controller data
	c.LoadControllerData()

	// Verifico se il player si trova in determinati stati non consentiti
	// e che quindi non permettano l'init del controller richiamato
	if inStateBlocked := c.InStatesBlocker(); inStateBlocked {
		return false
	}

	// Setto controller corrente nella cache
	helpers.SetCurrentControllerCache(c.Player.ID, c.CurrentState.Controller)

	// Carico payload
	if c.CurrentState.Stage, err = helpers.GetControllerCacheData(c.Player.ID, c.CurrentState.Controller, &c.CurrentState.Payload); err != nil {
		panic(err)
	}

	// Verifico se esistono condizioni per cambiare stato o uscire
	if c.BackTo(c.Configurations.ControllerBack.FromStage, c.Configurations.ControllerBack.To) {
		return false
	}

	return true
}

// Carico controller data
func (c *Controller) LoadControllerData() {
	var err error

	// Recupero stato utente
	var rGetActivePlayerStates *pb.GetActivePlayerStatesResponse
	if rGetActivePlayerStates, err = services.NnSDK.GetActivePlayerStates(helpers.NewContext(1), &pb.GetActivePlayerStatesRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		panic(err)
	}

	// Recupero stats utente
	var rGetPlayerStats *pb.GetPlayerStatsResponse
	if rGetPlayerStats, err = services.NnSDK.GetPlayerStats(helpers.NewContext(1), &pb.GetPlayerStatsRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		panic(err)
	}

	c.Data.PlayerActiveStates = rGetActivePlayerStates.GetStates()
	c.Data.PlayerStats = rGetPlayerStats.GetPlayerStats()
}

// Validate - Metodo comune per mandare messaggio di validazione
func (c *Controller) Validate() {
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

	if _, err := services.SendMessage(validatorMsg); err != nil {
		panic(err)
	}
}

// Breaking - Metodo che permette di verificare se si vogliono fare
// delle azioni che permetteranno di concludere
func (c *Controller) BackTo(canBackFromStage int32, controller ControllerInterface) (backed bool) {
	// Verifico se è effetivamente un messaggio di testo e non una callback
	if c.Update.Message != nil {
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.back") {
			if c.CurrentState.Controller != "" {
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
			helpers.DelCurrentControllerCache(c.Player.ID)
			helpers.DelControllerCacheData(c.Player.ID, c.CurrentState.Controller)

			new(MenuController).Handle(c.Player, c.Update)
			backed = true
			return
		}

		// Abbandona - chiude definitivamente cancellando anche lo stato
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.clears") ||
			c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.breaker.more") {
			if !c.Data.PlayerStats.GetDead() {
				// Cancello stato da cache
				helpers.DelCurrentControllerCache(c.Player.ID)
				helpers.DelControllerCacheData(c.Player.ID, c.CurrentState.Controller)

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
			helpers.DelCurrentControllerCache(c.Player.ID)

			// Call menu controller
			new(MenuController).Handle(c.Player, c.Update)

			backed = true
			return
		}
	}

	return
}

// Completing - Metodo per settare il completamento di uno stato
func (c *Controller) Completing(payload interface{}) (err error) {
	// Aggiorno stato controller
	if err = helpers.SetControllerCacheData(c.Player.ID, c.CurrentState.Controller, c.CurrentState.Stage, payload); err != nil {
		return
	}

	// Verifico se lo stato è completato chiudo
	if c.CurrentState.Completed {
		// Cancello stato dalla memoria
		helpers.DelCurrentControllerCache(c.Player.ID)
		helpers.DelControllerCacheData(c.Player.ID, c.CurrentState.Controller)

		// Call menu controller
		if c.Configurations.ControllerBack.To != nil {
			c.Update.Message.Text = ""
			c.Configurations.ControllerBack.To.Handle(c.Player, c.Update)
			return
		}

		new(MenuController).Handle(c.Player, c.Update)
		return
	}

	// Verifico se si vuole forzare il menu
	if c.ForceBackTo {
		// Cancello stato dalla memoria
		helpers.DelCurrentControllerCache(c.Player.ID)

		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update)
	}

	return
}

// InStatesBlocker
// Certi controller non possono essere eseguiti se il player si trova in determinati stati.
// Ogni controller ha la possibilità nell'handle di passare la lista di rotte bloccanti per esso.
func (c *Controller) InStatesBlocker() (inStates bool) {
	// Certi controller non devono subire la cancellazione degli stati
	// perchè magari hanno logiche particolari o lo gestiscono a loro modo
	for _, state := range c.Data.PlayerActiveStates {
		// Verifico se non fa parte dello stesso padre e che lo stato non sia completato
		if !state.Completed {
			for _, blockState := range c.Configurations.ControllerBlocked {
				if helpers.Trans(c.Player.Language.Slug, state.Controller) == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("route.%s", blockState)) {
					msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "valodator.controller.blocked"))
					if _, err := services.SendMessage(msg); err != nil {
						panic(err)
					}

					return true
				}
			}
		}
	}

	return false
}

// InTutorial - Metodo che semplifica il controllo se il player si trova dentro un tutorial
func (c *Controller) InTutorial() bool {
	for _, state := range c.Data.PlayerActiveStates {
		if state.Controller == "route.tutorial" {
			return true
		}
	}

	return false
}
