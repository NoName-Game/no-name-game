package controllers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipRepairsController
// ====================================
type ShipRepairsController struct {
	BaseController
	Payload struct {
		Ship              nnsdk.Ship
		QuantityResources int
		RepairTime        int
		TypeResources     string
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipRepairsController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	var playerStateProvider providers.PlayerStateProvider

	c.Controller = "route.ship.repairs"
	c.Player = player
	c.Update = update

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

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
		validatorMsg.ParseMode = "markdown"
		validatorMsg.ReplyMarkup = c.Validation.ReplyKeyboard

		_, err = services.SendMessage(validatorMsg)
		if err != nil {
			panic(err)
		}

		return
	}

	// Ok! Run!
	err = c.Stage()
	if err != nil {
		panic(err)
	}

	// Aggiorno stato finale
	payloadUpdated, _ := json.Marshal(c.Payload)
	c.State.Payload = string(payloadUpdated)
	c.State, err = playerStateProvider.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *ShipRepairsController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	switch c.State.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.repairs.start") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
					),
				),
			)

			return true, err
		}

		return false, err
	case 2:
		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"ship.repairs.wait",
			c.State.FinishAt.Format("15:04:05 01/02"),
		)

		c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
				),
			),
		)

		// Verifico se ha finito il crafting
		if time.Now().After(c.State.FinishAt) {
			return false, err
		}
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *ShipRepairsController) Stage() (err error) {
	var playerProvider providers.PlayerProvider
	var shipProvider providers.ShipProvider
	var resourceProvider providers.ResourceProvider

	switch c.State.Stage {

	// In questo riporto al player le risorse e tempistiche necessarie alla riparazione della nave
	case 0:
		// Recupero nave player equipaggiata
		var playerShips nnsdk.Ships
		playerShips, err = playerProvider.GetPlayerShips(c.Player, true)
		if err != nil {
			return err
		}

		// TODO: verificare, dovrebbe recuperarne solo una
		// Recupero name del player
		var playerShip nnsdk.Ship
		playerShip = playerShips[0]

		// Recupero informazioni nave da riparare
		var repairInfo nnsdk.ShipRepairInfoResponse
		repairInfo, err = shipProvider.GetShipRepairInfo(playerShip)
		if err != nil {
			return err
		}

		// Verifico se effettivamente la nave è da riparare
		var shipRecap string
		shipRecap = helpers.Trans(c.Player.Language.Slug, "ship.repairs.info")
		if repairInfo.NeedRepairs {
			shipRecap += fmt.Sprintf("🔧 %v/100%% (%s)\n%s\n%s ",
				playerShip.ShipStats.Integrity, helpers.Trans(c.Player.Language.Slug, "integrity"),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.time", repairInfo.RepairTime),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.quantity_resources", repairInfo.QuantityResources, repairInfo.TypeResources),
			)
		} else {
			shipRecap = helpers.Trans(c.Player.Language.Slug, "ship.repairs.dont_need")
		}

		// Aggiongo bottone start riparazione
		var keyboardRow [][]tgbotapi.KeyboardButton
		if repairInfo.NeedRepairs {
			newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "ship.repairs.start"),
				),
			)
			keyboardRow = append(keyboardRow, newKeyboardRow)
		}

		// Clear and exit
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
		))

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, shipRecap)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRow,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.Payload.Ship = playerShip
		c.Payload.QuantityResources = repairInfo.QuantityResources
		c.Payload.RepairTime = repairInfo.RepairTime
		c.Payload.TypeResources = repairInfo.TypeResources
		c.State.Stage = 1

	// In questo stage avvio effettivamente la riparzione
	case 1:
		// Avvio riparazione nave
		var resourcesUsed []nnsdk.ShipRepairStartResponse
		resourcesUsed, err = shipProvider.StartShipRepair(c.Payload.Ship)
		if err != nil && strings.Contains(err.Error(), "not enough resource quantities") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := services.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.not_enough_resource"),
			)
			_, err = services.SendMessage(errorMsg)
			if err != nil {
				return err
			}

			return
		}

		// Se tutto ok mostro le risorse che vengono consumate per la riparazione
		var recapResourceUsed string
		recapResourceUsed = helpers.Trans(c.Player.Language.Slug, "ship.repairs.used_resources")
		for _, resourceUsed := range resourcesUsed {
			var resource nnsdk.Resource
			resource, err = resourceProvider.GetResourceByID(resourceUsed.ResourceID)
			if err != nil {
				return err
			}

			recapResourceUsed += fmt.Sprintf("\n- %s x %v", resource.Name, resourceUsed.Quantity)
		}

		// Setto timer recuperato dalla chiamata delle info
		c.State.FinishAt = helpers.GetEndTime(0, int(c.Payload.RepairTime), 0)

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s \n\n%s",
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.reparing", c.State.FinishAt.Format("15:04:05")),
				recapResourceUsed,
			),
		)

		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.State.ToNotify = helpers.SetTrue()
		c.State.Stage = 2
	case 2:
		// Fine riparazione
		err = shipProvider.EndShipRepair(c.Payload.Ship)
		if err != nil {
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.repairs.reparing.finish"),
		)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
				),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.State.Completed = helpers.SetTrue()
	}

	return
}
