package controllers

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipRepairsController
// ====================================
type ShipRepairsController struct {
	BaseController
	Payload struct {
		Ship              *pb.Ship
		QuantityResources int32
		RepairTime        int32
		TypeResources     string
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipRepairsController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller: "route.ship.repairs",
		ControllerBack: ControllerBack{
			To:        &ShipController{},
			FromStage: 1,
		},
		Payload: c.Payload,
	}) {
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)

	// Validate
	var hasError bool
	if hasError = c.Validator(); hasError {
		c.Validate()
		return
	}

	// Ok! Run!
	if err = c.Stage(); err != nil {
		panic(err)
	}

	// Completo progressione
	if err = c.Completing(c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *ShipRepairsController) Validator() (hasErrors bool) {
	var err error
	switch c.PlayerData.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "ship.repairs.start") {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
					),
				),
			)

			return true
		}

		return false
	case 2:
		var finishAt time.Time
		finishAt, err = ptypes.Timestamp(c.PlayerData.CurrentState.FinishAt)
		if err != nil {
			panic(err)
		}

		c.Validation.Message = helpers.Trans(
			c.Player.Language.Slug,
			"ship.repairs.wait",
			finishAt.Format("15:04:05 01/02"),
		)

		// Aggiungo anche abbandona
		c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.continue"),
				),
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
				),
			),
		)

		// Verifico se ha finito il crafting
		if time.Now().After(finishAt) {
			return false
		}
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *ShipRepairsController) Stage() (err error) {
	switch c.PlayerData.CurrentState.Stage {

	// In questo riporto al player le risorse e tempistiche necessarie alla riparazione della nave
	case 0:
		// Recupero nave player equipaggiata
		rGetPlayerShipEquipped, err := services.NnSDK.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		// Recupero informazioni nave da riparare
		rGetShipRepairInfo, err := services.NnSDK.GetShipRepairInfo(helpers.NewContext(1), &pb.GetShipRepairInfoRequest{
			Ship: rGetPlayerShipEquipped.GetShip(),
		})
		if err != nil {
			return err
		}

		// Verifico se effettivamente la nave è da riparare
		var shipRecap string
		shipRecap = helpers.Trans(c.Player.Language.Slug, "ship.repairs.info")
		if rGetShipRepairInfo.GetNeedRepairs() {
			shipRecap += fmt.Sprintf("🔧 %v/100%% (%s)\n%s\n%s ",
				rGetPlayerShipEquipped.GetShip().GetShipStats().GetIntegrity(), helpers.Trans(c.Player.Language.Slug, "integrity"),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.time", rGetShipRepairInfo.GetRepairTime()),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.quantity_resources", rGetShipRepairInfo.GetQuantityResources(), rGetShipRepairInfo.GetTypeResources()),
			)
		} else {
			shipRecap = helpers.Trans(c.Player.Language.Slug, "ship.repairs.dont_need")
		}

		// Aggiongo bottone start riparazione
		var keyboardRow [][]tgbotapi.KeyboardButton
		if rGetShipRepairInfo.GetNeedRepairs() {
			newKeyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "ship.repairs.start"),
				),
			)
			keyboardRow = append(keyboardRow, newKeyboardRow)
		}

		// Clear and exit
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
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
		c.Payload.Ship = rGetPlayerShipEquipped.GetShip()
		c.Payload.QuantityResources = rGetShipRepairInfo.GetQuantityResources()
		c.Payload.RepairTime = rGetShipRepairInfo.GetRepairTime()
		c.Payload.TypeResources = rGetShipRepairInfo.GetTypeResources()
		c.PlayerData.CurrentState.Stage = 1

	// In questo stage avvio effettivamente la riparzione
	case 1:
		// Avvio riparazione nave
		rStartShipRepair, err := services.NnSDK.StartShipRepair(helpers.NewContext(1), &pb.StartShipRepairRequest{
			Ship: c.Payload.Ship,
		})

		if err != nil && strings.Contains(err.Error(), "not enough resource quantities") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := services.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.not_enough_resource"),
			)
			_, err = services.SendMessage(errorMsg)
			if err != nil {
				return err
			}

			return err
		}

		// Se tutto ok mostro le risorse che vengono consumate per la riparazione
		var recapResourceUsed string
		recapResourceUsed = helpers.Trans(c.Player.Language.Slug, "ship.repairs.used_resources")
		for _, resourceUsed := range rStartShipRepair.GetStartShipRepair() {
			var rGetResourceByID *pb.GetResourceByIDResponse
			rGetResourceByID, err = services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: resourceUsed.ResourceID,
			})
			if err != nil {
				return err
			}

			recapResourceUsed += fmt.Sprintf("\n- %s x %v", rGetResourceByID.GetResource().GetName(), resourceUsed.Quantity)
		}

		// Setto timer recuperato dalla chiamata delle info
		finishTime := helpers.GetEndTime(0, int(c.Payload.RepairTime), 0)
		c.PlayerData.CurrentState.FinishAt, _ = ptypes.TimestampProto(finishTime)

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s \n\n%s",
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.reparing", finishTime.Format("15:04:05")),
				recapResourceUsed,
			),
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.PlayerData.CurrentState.ToNotify = true
		c.PlayerData.CurrentState.Stage = 2
		c.ForceBackTo = true
	case 2:
		// Fine riparazione
		_, err := services.NnSDK.EndShipRepair(helpers.NewContext(1), &pb.EndShipRepairRequest{
			Ship: c.Payload.Ship,
		})
		if err != nil {
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.repairs.reparing.finish"),
		)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.PlayerData.CurrentState.Completed = true
	}

	return
}
