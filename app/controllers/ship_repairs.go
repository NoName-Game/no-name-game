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
		ShipID     uint32
		RepairType pb.StartShipRepairRequest_RapairTypeEnum
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

	// Carico payload
	if err = helpers.GetPayloadController(c.Player.ID, c.CurrentState.Controller, &c.Payload); err != nil {
		panic(err)
	}

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
	if err = c.Completing(&c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *ShipRepairsController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		// Recupero nave player equipaggiata
		var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
		if rGetPlayerShipEquipped, err = services.NnSDK.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			panic(err)
		}

		// Recupero informazioni nave da riparare
		var rGetShipRepairInfo *pb.GetShipRepairInfoResponse
		if rGetShipRepairInfo, err = services.NnSDK.GetShipRepairInfo(helpers.NewContext(1), &pb.GetShipRepairInfoRequest{
			ShipID: rGetPlayerShipEquipped.GetShip().GetID(),
		}); err != nil {
			panic(err)
		}

		if !rGetShipRepairInfo.GetNeedRepairs() {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ship.repairs.dont_need")
			return true
		}

		return false
	case 1:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "ship.repairs.start_partial") {
			c.Payload.RepairType = pb.StartShipRepairRequest_PARTIAL
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "ship.repairs.start_full") {
			c.Payload.RepairType = pb.StartShipRepairRequest_FULL
			return false
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
				),
			),
		)

		return true

	case 2:
		var rCheckShipRepair *pb.CheckShipRepairResponse
		if rCheckShipRepair, err = services.NnSDK.CheckShipRepair(helpers.NewContext(1), &pb.CheckShipRepairRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			panic(err)
		}

		// Il crafter sta già portando a terminre un lavoro per questo player
		if !rCheckShipRepair.GetFinishRepairing() {
			var finishAt time.Time
			if finishAt, err = ptypes.Timestamp(rCheckShipRepair.GetRepairingEndTime()); err != nil {
				panic(err)
			}

			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"ship.repairs.wait",
				finishAt.Format("15:04:05"),
			)

			return true
		}

		return false
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *ShipRepairsController) Stage() (err error) {
	switch c.CurrentState.Stage {

	// In questo riporto al player le risorse e tempistiche necessarie alla riparazione della nave
	case 0:
		// Recupero nave player equipaggiata
		var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
		if rGetPlayerShipEquipped, err = services.NnSDK.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			return err
		}

		// Recupero informazioni nave da riparare
		var rGetShipRepairInfo *pb.GetShipRepairInfoResponse
		if rGetShipRepairInfo, err = services.NnSDK.GetShipRepairInfo(helpers.NewContext(1), &pb.GetShipRepairInfoRequest{
			ShipID: rGetPlayerShipEquipped.GetShip().GetID(),
		}); err != nil {
			return err
		}

		var shipRecap string
		shipRecap = helpers.Trans(c.Player.Language.Slug, "ship.repairs.info")
		if rGetShipRepairInfo.GetNeedRepairs() {
			// Mostro Partial
			shipRecap += fmt.Sprintf("%s\n🔧 %v%% ➡️ *%v%%* (%s)\n%s\n%s\n\n",
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.partial"),
				rGetPlayerShipEquipped.GetShip().GetShipStats().GetIntegrity(), rGetShipRepairInfo.GetPartial().GetIntegrity(), helpers.Trans(c.Player.Language.Slug, "integrity"),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.time", rGetShipRepairInfo.GetPartial().GetRepairTime()),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.quantity_resources", rGetShipRepairInfo.GetPartial().GetQuantityResources()),
			)

			// Mostro Full
			shipRecap += fmt.Sprintf("%s\n🔧 %v%% ➡️ *100%%* (%s)\n%s\n%s ",
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.full"),
				rGetPlayerShipEquipped.GetShip().GetShipStats().GetIntegrity(), helpers.Trans(c.Player.Language.Slug, "integrity"),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.time", rGetShipRepairInfo.GetFull().GetRepairTime()),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.quantity_resources", rGetShipRepairInfo.GetFull().GetQuantityResources()),
			)
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, shipRecap)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "ship.repairs.start_partial")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "ship.repairs.start_full")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	// In questo stage avvio effettivamente la riparzione
	case 1:
		// Avvio riparazione nave
		var rStartShipRepair *pb.StartShipRepairResponse
		rStartShipRepair, err = services.NnSDK.StartShipRepair(helpers.NewContext(1), &pb.StartShipRepairRequest{
			PlayerID:   c.Player.ID,
			RapairType: c.Payload.RepairType,
		})

		if err != nil && strings.Contains(err.Error(), "not enough resource quantities") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := services.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.not_enough_resource"),
			)
			if _, err = services.SendMessage(errorMsg); err != nil {
				return err
			}

			return err
		}

		// Se tutto ok mostro le risorse che vengono consumate per la riparazione
		var recapResourceUsed string
		recapResourceUsed = helpers.Trans(c.Player.Language.Slug, "ship.repairs.used_resources")
		for _, resourceUsed := range rStartShipRepair.GetStartShipRepair() {
			var rGetResourceByID *pb.GetResourceByIDResponse
			if rGetResourceByID, err = services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: resourceUsed.ResourceID,
			}); err != nil {
				return err
			}

			recapResourceUsed += fmt.Sprintf("\n- *%v* x %s (%s)",
				resourceUsed.Quantity,
				rGetResourceByID.GetResource().GetName(), rGetResourceByID.GetResource().GetRarity().GetSlug(),
			)
		}

		// Recupero orario fine riparazione
		var finishAt time.Time
		if finishAt, err = ptypes.Timestamp(rStartShipRepair.GetRepairingEndTime()); err != nil {
			return
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s \n\n%s",
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.reparing", finishAt.Format("15:04:05")),
				recapResourceUsed,
			),
		)
		msg.ParseMode = "markdown"
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
		c.ForceBackTo = true
	case 2:
		// Fine riparazione
		if _, err := services.NnSDK.EndShipRepair(helpers.NewContext(1), &pb.EndShipRepairRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			return err
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.repairs.reparing.finish"),
		)
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
