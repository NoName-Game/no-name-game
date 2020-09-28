package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetTitanController
// ====================================
type SafePlanetTitanController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetTitanController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.titan",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 1,
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

	// Completo progressione
	c.Completing(nil)
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetTitanController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	case 1:
		// Recupero quali titani sono stati scoperti e quindi raggiungibili
		var rTitanDiscovered *pb.TitanDiscoveredResponse
		if rTitanDiscovered, err = config.App.Server.Connection.TitanDiscovered(helpers.NewContext(1), &pb.TitanDiscoveredRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		// Verifico sei il player ha passato il nome di un titano valido
		if len(rTitanDiscovered.GetTitans()) > 0 {
			for _, titan := range rTitanDiscovered.GetTitans() {
				if c.Update.Message.Text == titan.GetName() {
					return false
				}
			}
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetTitanController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		var restsRecap string
		restsRecap = helpers.Trans(c.Player.Language.Slug, "route.safeplanet.titan.info")
		var keyboardRow [][]tgbotapi.KeyboardButton

		// Recupero quali titani sono stati scoperti e quindi raggiungibili
		var rTitanDiscovered *pb.TitanDiscoveredResponse
		if rTitanDiscovered, err = config.App.Server.Connection.TitanDiscovered(helpers.NewContext(1), &pb.TitanDiscoveredRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		// Se sono stati trovati dei tiani costruisco keyboard
		if len(rTitanDiscovered.GetTitans()) > 0 {
			restsRecap += helpers.Trans(c.Player.Language.Slug, "route.safeplanet.titan.choice")
			for _, titan := range rTitanDiscovered.GetTitans() {
				newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						titan.GetName(),
					),
				)
				keyboardRow = append(keyboardRow, newKeyboardRow)
			}
		} else {
			restsRecap += helpers.Trans(c.Player.Language.Slug, "route.safeplanet.titan.no_titans_founded")
		}

		// Aggiungo torna indietro
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.Player.ChatID, restsRecap)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	// In questo stage avvio effettivamente il riposo
	case 1:
		// Recupero pianeta da titano
		var rGetTitanByName *pb.GetTitanByNameResponse
		if rGetTitanByName, err = config.App.Server.Connection.GetTitanByName(helpers.NewContext(1), &pb.GetTitanByNameRequest{
			Name: c.Update.Message.Text,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiunto nuova posizione al player
		if _, err = config.App.Server.Connection.CreatePlayerPosition(helpers.NewContext(1), &pb.CreatePlayerPositionRequest{
			PlayerID: c.Player.ID,
			PlanetID: rGetTitanByName.GetTitan().GetPlanetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "route.safeplanet.titan.teleport"),
		)

		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
		c.Configurations.ControllerBack.To = &MenuController{}
	}

	return
}
