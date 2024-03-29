package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"
)

// ====================================
// SafePlanetTitanController
// ====================================
type SafePlanetTitanController struct {
	Controller
}

func (c *SafePlanetTitanController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.titan",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 1,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetTitanController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Validate
	if c.Validator() {
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
	case 0:
		// ##################################################################################################
		// Verifico che la nave equipaggiata non sia in riparazione
		// ##################################################################################################
		for _, state := range c.Data.PlayerActiveStates {
			if state.GetController() == "route.safeplanet.hangar.repair" {
				var repairingData struct {
					ShipID uint32
				}

				_ = json.Unmarshal([]byte(state.GetPayload()), &repairingData)

				// Recupero nave attualemente attiva
				var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
				if rGetPlayerShipEquipped, err = config.App.Server.Connection.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
					PlayerID: c.Player.GetID(),
				}); err != nil {
					c.Logger.Panic(err)
				}

				if rGetPlayerShipEquipped.GetShip().GetID() == repairingData.ShipID {
					c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.controller.blocked")
					return true
				}
			}
		}
	// ##################################################################################################
	// Verifico sei il player ha passato il nome di un titano valido
	// ##################################################################################################
	case 1:
		var rTitanDiscovered *pb.TitanDiscoveredResponse
		if rTitanDiscovered, err = config.App.Server.Connection.TitanDiscovered(helpers.NewContext(1), &pb.TitanDiscoveredRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		if len(rTitanDiscovered.GetTitans()) > 0 {
			titanMsg := strings.Split(c.Update.Message.Text, " ")[1]
			for _, titan := range rTitanDiscovered.GetTitans() {
				if titanMsg == titan.GetName() {
					return false
				}
			}
		}

		return true
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetTitanController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		var restsRecap string
		restsRecap = helpers.Trans(c.Player.Language.Slug, "safeplanet.titan.info")
		var keyboardRow [][]tgbotapi.KeyboardButton
		// Recupero quali titani sono stati scoperti e quindi raggiungibili
		var rTitanDiscovered *pb.TitanDiscoveredResponse
		if rTitanDiscovered, err = config.App.Server.Connection.TitanDiscovered(helpers.NewContext(1), &pb.TitanDiscoveredRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		// Se sono stati trovati dei tiani costruisco keyboard
		if len(rTitanDiscovered.GetTitans()) > 0 {
			restsRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.titan.choice")
			for _, titan := range rTitanDiscovered.GetTitans() {
				var titanStatus = "⚔️"
				if titan.GetKilledAt() != nil {
					titanStatus = "☠️"
				}

				newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						fmt.Sprintf("%s %s 💫 %v 🎖%v",
							titanStatus,
							titan.GetName(),
							titan.GetPlanetSystemID(),
							titan.GetPlanetSystemID()*15,
						),
					),
				)
				keyboardRow = append(keyboardRow, newKeyboardRow)
			}
		} else {
			restsRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.titan.no_titans_founded")
		}

		// Aggiungo torna indietro
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.Player.ChatID, restsRecap)
		msg.ParseMode = tgbotapi.ModeHTML
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
		titanMsg := strings.Split(c.Update.Message.Text, " ")[1]

		// Recupero pianeta da titano
		var rGetTitanByName *pb.GetTitanByNameResponse
		if rGetTitanByName, err = config.App.Server.Connection.GetTitanByName(helpers.NewContext(1), &pb.GetTitanByNameRequest{
			Name: titanMsg,
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
		msg := helpers.NewMessage(c.ChatID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.titan.teleport"),
		)

		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Forzo cancellazione posizione player in cache
		_ = helpers.DelPlayerPlanetPositionInCache(c.Player.GetID())

		// Completo lo stato
		c.CurrentState.Completed = true
		c.Configurations.ControllerBack.To = &MenuController{}
	}

	return
}
