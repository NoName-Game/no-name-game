package controllers

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Controller - Intereffaccia base di tutti i controller
type ControllerInterface interface {
	Handle(*pb.Player, tgbotapi.Update)
	Validator() bool
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
		PlayerActiveStates    []*pb.PlayerActivity
		PlayerCurrentPosition *pb.Planet
		// PlayerStats        *pb.PlayerStats
	}
	ForceBackTo    bool
	Configurations ControllerConfigurations
	Logger         *logrus.Entry
}

type ControllerCurrentState struct {
	Controller string
	Stage      int32
	Completed  bool
	Payload    interface{}
}

type ControllerConfigurations struct {
	CustomBreaker     []string
	ControllerBlocked []string
	ControllerBack    ControllerBack
	PlanetType        []string
}

type ControllerBack struct {
	To        ControllerInterface
	FromStage int32
}

func (c *Controller) InitController(controller Controller) bool {
	var err error

	// Carico configurazione controller
	*c = controller

	// Carico infomazioni per logger
	c.SetLoggerData()

	// Carico controller data
	if err = c.LoadControllerData(); err != nil {
		c.Logger.Panicf("cant load controller data: %s", err.Error())
	}

	// Verifico che il player si trovi nel pianeta consentito
	if correctPlanet := c.InCorrectPlanet(); !correctPlanet {
		return false
	}

	// Verifico se il player si trova in determinati stati non consentiti
	// e che quindi non permettano l'init del controller richiamato
	if inStateBlocked := c.InStatesBlocker(); inStateBlocked {
		return false
	}

	// Setto controller corrente nella cache
	helpers.SetCurrentControllerCache(c.Player.ID, c.CurrentState.Controller)

	// Carico payload e infomazioni controller
	if c.CurrentState.Stage, err = helpers.GetControllerCacheData(c.Player.ID, c.CurrentState.Controller, &c.CurrentState.Payload); err != nil {
		c.Logger.Panicf("cant get stage and paylaod controller data: %s", err.Error())
	}

	// Verifico se esistono condizioni per cambiare stato o uscire
	if c.BackTo(c.Configurations.ControllerBack.FromStage, c.Configurations.ControllerBack.To) {
		return false
	}

	return true
}

func (c *Controller) SetLoggerData() {
	var message string
	if c.Update.Message != nil {
		message = c.Update.Message.Text
	} else if c.Update.CallbackQuery != nil {
		message = c.Update.CallbackQuery.Data
	}

	c.Logger = logrus.WithFields(logrus.Fields{
		"controller": c.CurrentState.Controller,
		"stage":      c.CurrentState.Stage,
		"player":     c.Player.ID,
		"message":    message,
	})
}

// Carico controller data
func (c *Controller) LoadControllerData() (err error) {
	// Recupero stato utente
	var rGetActivePlayerActivities *pb.GetActivePlayerActivitiesResponse
	if rGetActivePlayerActivities, err = config.App.Server.Connection.GetActivePlayerActivities(helpers.NewContext(1), &pb.GetActivePlayerActivitiesRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		return err
	}

	c.Data.PlayerActiveStates = rGetActivePlayerActivities.GetActivities()

	// Recupero posizione player.
	var currentPosition *pb.Planet
	if currentPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
		c.Logger.Panic(err)
	}

	c.Data.PlayerCurrentPosition = currentPosition
	return
}

func (c *Controller) RegisterError(err error) {
	// Registro errore su sentry
	c.Logger.Error(err)

	// Invio il messaggio in caso di errore e chiudo
	validatorMsg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "validator.error"))
	validatorMsg.ParseMode = tgbotapi.ModeMarkdown
	validatorMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		),
	)

	if _, err := helpers.SendMessage(validatorMsg); err != nil {
		panic(err)
	}
}

// Validate - Metodo comune per mandare messaggio di validazione
func (c *Controller) Validate() {
	// Se non ha un messaggio particolare allora setto configurazione di default
	if c.Validation.Message == "" {
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
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
	validatorMsg := helpers.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
	validatorMsg.ParseMode = tgbotapi.ModeMarkdown
	if c.Validation.ReplyKeyboard.Keyboard != nil {
		validatorMsg.ReplyMarkup = c.Validation.ReplyKeyboard
	}

	if _, err := helpers.SendMessage(validatorMsg); err != nil {
		panic(err)
	}
}

// Breaking - Metodo che permette di verificare se si vogliono fare
// delle azioni che permetteranno di concludere
func (c *Controller) BackTo(canBackFromStage int32, controller ControllerInterface) (backed bool) {
	if c.Update.Message != nil {
		switch c.Update.Message.Text {
		// Questo braker forza il player a tornare al menu precedente senza cancellare lo stato
		case helpers.Trans(c.Player.Language.Slug, "route.breaker.back"):
			if c.CurrentState.Stage > canBackFromStage {
				c.CurrentState.Stage = 0
				return
			}

			// Cancello stato dalla memoria
			helpers.DelCurrentControllerCache(c.Player.ID)

			// Se è stato settato un controller esco
			if controller != nil {
				// Rimuovo testo messaggio
				c.Update.Message.Text = ""
				controller.Handle(c.Player, c.Update)
				backed = true
				return
			}

			new(MenuController).Handle(c.Player, c.Update)
			backed = true
			return

		// Questo braker forza il player a tornare al menù precedente
		case helpers.Trans(c.Player.Language.Slug, "route.breaker.menu"):
			// Cancello stato dalla memoria
			helpers.DelCurrentControllerCache(c.Player.ID)
			_ = helpers.DelControllerCacheData(c.Player.ID, c.CurrentState.Controller)

			if controller != nil {
				// Rimuovo testo messaggio
				c.Update.Message.Text = ""
				controller.Handle(c.Player, c.Update)
				backed = true
				return
			}

			new(MenuController).Handle(c.Player, c.Update)
			backed = true
			return

		// Questo braker forza il player a tornare al menù principale cancellando l'attività
		case helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"):
			if !c.Player.GetDead() {
				// Cancello stato da cache
				helpers.DelCurrentControllerCache(c.Player.ID)
				_ = helpers.DelControllerCacheData(c.Player.ID, c.CurrentState.Controller)

				// Call menu controller
				new(MenuController).Handle(c.Player, c.Update)

				backed = true
				return
			}
		case helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"):
			if !c.Player.GetDead() {
				// Cancello stato da cache
				helpers.DelCurrentControllerCache(c.Player.ID)
				_ = helpers.DelControllerCacheData(c.Player.ID, c.CurrentState.Controller)

				// Call menu controller
				new(MenuController).Handle(c.Player, c.Update)

				backed = true
				return
			}
		}

		// Continua - mantiene lo stato attivo ma ti forza a tornare al menù
		// usato principalemente per notificare che esiste già un'attività in corso (Es. Missione)
		c.Configurations.CustomBreaker = append(c.Configurations.CustomBreaker, "route.breaker.continue")
		if helpers.MessageInCustomBreakers(c.Update.Message.Text, c.Player.Language.Slug, c.Configurations.CustomBreaker) {
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
func (c *Controller) Completing(payload interface{}) {
	var err error

	// Aggiorno stato controller
	if err = helpers.SetControllerCacheData(c.Player.ID, c.CurrentState.Controller, c.CurrentState.Stage, payload); err != nil {
		c.Logger.Panicf("cant set controller cache data: %s", err.Error())
		return
	}

	// Verifico se lo stato è completato chiudo
	if c.CurrentState.Completed {
		// Cancello stato dalla memoria
		helpers.DelCurrentControllerCache(c.Player.ID)
		_ = helpers.DelControllerCacheData(c.Player.ID, c.CurrentState.Controller)

		// Call menu controller
		if c.Configurations.ControllerBack.To != nil {
			if c.Update.Message != nil {
				c.Update.Message.Text = ""
				c.Configurations.ControllerBack.To.Handle(c.Player, c.Update)
				return
			}
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

		// Verifico sie il player sta dormendo, in questo caso non può effettuare nessuna azione
		if state.GetController() == "route.ship.rests" && state.GetStage() > 1 {
			// Se un azione è diversa dal risvegliati
			if helpers.Trans(c.Player.Language.Slug, "ship.rests.wakeup") != c.Update.Message.Text {
				msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "ship.rests.validator.need_to_wakeup"))
				if _, err := helpers.SendMessage(msg); err != nil {
					panic(err)
				}

				return true
			}
		}

		// Verifico se ci sono stati particolare bloccati
		for _, blockState := range c.Configurations.ControllerBlocked {
			if helpers.Trans(c.Player.Language.Slug, state.Controller) == helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("route.%s", blockState)) {
				msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "validator.controller.blocked"))
				if _, err := helpers.SendMessage(msg); err != nil {
					panic(err)
				}

				return true
			}
		}
	}

	return false
}

// InCorrectPlanet
// Alcuni controller hanno azioni richiamabili solo da determinati pianeti (sicuro, titano, normale).
// Effettuare il controllo e ritornare TRUE se l'azione può proseguire.
// Tipi di pianeta consentiti: {safe, titan, default}
func (c *Controller) InCorrectPlanet() (correctPlanet bool) {
	// Se l'array è vuoto tutti i pianeti sono consentiti. Skippo il controllo.
	if len(c.Configurations.PlanetType) == 0 {
		return true
	}

	inSafe := c.CheckInSafePlanet(c.Data.PlayerCurrentPosition)
	inTitan, _ := c.CheckInTitanPlanet(c.Data.PlayerCurrentPosition)
	inDarkMerchantPlanet := c.CheckInDarkMerchantPlanet(c.Data.PlayerCurrentPosition)

	// Ciclo fra i tipi di pianeta consentito
	for _, planetType := range c.Configurations.PlanetType {
		switch planetType {
		case "safe":
			return inSafe
		case "titan":
			return inTitan
		case "darkMerchant":
			return inDarkMerchantPlanet
		case "default":
			// Se entrambi i valori sono falsi allora è un pianeta classico
			if !(inSafe || inTitan) {
				return true
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

// CheckInTitanPlanet
// Verifico se il player si trova su un pianeta sicuro
func (c *Controller) CheckInTitanPlanet(position *pb.Planet) (inTitanPlanet bool, titan *pb.Titan) {
	// Verifico se il pianeta corrente è occupato da un titano
	var rGetTitanByPlanetID *pb.GetTitanByPlanetIDResponse
	rGetTitanByPlanetID, _ = config.App.Server.Connection.GetTitanByPlanetID(helpers.NewContext(1), &pb.GetTitanByPlanetIDRequest{
		PlanetID: position.GetID(),
	})

	if rGetTitanByPlanetID.GetTitan().GetID() > 0 {
		return true, rGetTitanByPlanetID.GetTitan()
	}

	return false, nil
}

// CheckInDarkMerchantPlanet
// Verifico se il player si trova sul pianeta del mercante oscuro
func (c *Controller) CheckInDarkMerchantPlanet(position *pb.Planet) (inDarkMerchant bool) {
	var err error
	var rGetDarkMerchant *pb.GetDarkMerchantResponse
	if rGetDarkMerchant, err = config.App.Server.Connection.GetDarkMerchant(helpers.NewContext(1), &pb.GetDarkMerchantRequest{}); err != nil {
		c.Logger.Panic(err)
	}

	if rGetDarkMerchant.GetPlanetID() == position.ID {
		return true
	}

	return false
}

// CheckInSafePlanet
// Verifico se il player si trova su un pianeta sicuro
func (c *Controller) CheckInSafePlanet(position *pb.Planet) bool {
	return position.GetSafe()
}
