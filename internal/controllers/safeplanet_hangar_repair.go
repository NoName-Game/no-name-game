package controllers

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetHangarRepairController
// ====================================
type SafePlanetHangarRepairController struct {
	Controller
	Payload struct {
		ShipID     uint32
		RepairType pb.StartShipRepairRequest_RapairTypeEnum
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetHangarRepairController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.hangar.repair",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetHangarController{},
				FromStage: 1,
			},
		},
	}) {
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
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetHangarRepairController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se la nave attualmente equipaggiata dal player necessita di riparazioni
	// ##################################################################################################
	case 0:
		var err error
		var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
		if rGetPlayerShipEquipped, err = config.App.Server.Connection.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero informazioni nave da riparare
		var rGetShipRepairInfo *pb.GetShipRepairInfoResponse
		if rGetShipRepairInfo, err = config.App.Server.Connection.GetShipRepairInfo(helpers.NewContext(1), &pb.GetShipRepairInfoRequest{
			ShipID: rGetPlayerShipEquipped.GetShip().GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		if !rGetShipRepairInfo.GetNeedRepairs() {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.dont_need")
			return true
		}
	// ##################################################################################################
	// Verifico quale tipologia di riparazione vuuole effettuare il player
	// ##################################################################################################
	case 1:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.start_partial") {
			c.Payload.RepairType = pb.StartShipRepairRequest_PARTIAL
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.start_full") {
			c.Payload.RepairType = pb.StartShipRepairRequest_FULL
			return false
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
				),
			),
		)

		return true

	// ##################################################################################################
	// Verifico stato riparazione
	// ##################################################################################################
	case 2:
		var err error
		var rCheckShipRepair *pb.CheckShipRepairResponse
		if rCheckShipRepair, err = config.App.Server.Connection.CheckShipRepair(helpers.NewContext(1), &pb.CheckShipRepairRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Il crafter sta già portando a terminre un lavoro per questo player
		if !rCheckShipRepair.GetFinishRepairing() {
			var finishAt time.Time
			if finishAt, err = helpers.GetEndTime(rCheckShipRepair.GetRepairingEndTime(), c.Player); err != nil {
				c.Logger.Panic(err)
			}

			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"safeplanet.hangar.wait",
				finishAt.Format("15:04:05"),
			)

			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetHangarRepairController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// In questo riporto al player le risorse e tempistiche necessarie alla riparazione della nave
	case 0:
		// Recupero nave player equipaggiata
		var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
		if rGetPlayerShipEquipped, err = config.App.Server.Connection.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero informazioni nave da riparare
		var rGetShipRepairInfo *pb.GetShipRepairInfoResponse
		if rGetShipRepairInfo, err = config.App.Server.Connection.GetShipRepairInfo(helpers.NewContext(1), &pb.GetShipRepairInfoRequest{
			ShipID: rGetPlayerShipEquipped.GetShip().GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		var shipRecap string
		shipRecap = helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.info")
		if rGetShipRepairInfo.GetNeedRepairs() {
			// Mostro Partial
			shipRecap += fmt.Sprintf("%s\n🔧 %v%% ➡️ *%v%%* (%s)\n%s\n%s\n\n",
				helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.partial"),
				rGetPlayerShipEquipped.GetShip().GetIntegrity(), rGetShipRepairInfo.GetPartial().GetIntegrity(), helpers.Trans(c.Player.Language.Slug, "integrity"),
				helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.time", rGetShipRepairInfo.GetPartial().GetRepairTime()),
				helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.quantity_resources", rGetShipRepairInfo.GetPartial().GetQuantityResources()),
			)

			// Mostro Full
			shipRecap += fmt.Sprintf("%s\n🔧 %v%% ➡️ *100%%* (%s)\n%s\n%s ",
				helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.full"),
				rGetPlayerShipEquipped.GetShip().GetIntegrity(), helpers.Trans(c.Player.Language.Slug, "integrity"),
				helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.time", rGetShipRepairInfo.GetFull().GetRepairTime()),
				helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.quantity_resources", rGetShipRepairInfo.GetFull().GetQuantityResources()),
			)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, shipRecap)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.start_partial")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.start_full")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	// In questo stage avvio effettivamente la riparzione
	case 1:
		// Avvio riparazione nave
		var rStartShipRepair *pb.StartShipRepairResponse
		rStartShipRepair, err = config.App.Server.Connection.StartShipRepair(helpers.NewContext(1), &pb.StartShipRepairRequest{
			PlayerID:   c.Player.ID,
			RapairType: c.Payload.RepairType,
		})

		if err != nil && strings.Contains(err.Error(), "not enough resource quantities") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.not_enough_resource"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}
			return
		}

		// Se tutto ok mostro le risorse che vengono consumate per la riparazione
		var recapResourceUsed string
		recapResourceUsed = helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.used_resources")
		for _, resourceUsed := range rStartShipRepair.GetStartShipRepair() {
			var rGetResourceByID *pb.GetResourceByIDResponse
			if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: resourceUsed.ResourceID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			recapResourceUsed += fmt.Sprintf("\n- *%v* x %s (%s)",
				resourceUsed.Quantity,
				rGetResourceByID.GetResource().GetName(), rGetResourceByID.GetResource().GetRarity().GetSlug(),
			)
		}

		// Recupero orario fine riparazione
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rStartShipRepair.GetRepairingEndTime(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s \n\n%s",
				helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.reparing", finishAt.Format("15:04:05")),
				recapResourceUsed,
			),
		)
		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
		c.ForceBackTo = true
	case 2:
		// Fine riparazione
		if _, err := config.App.Server.Connection.EndShipRepair(helpers.NewContext(1), &pb.EndShipRepairRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.reparing.finish"),
		)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
