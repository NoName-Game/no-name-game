package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

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
func (c *ShipRepairsController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.ship.repairs",
		c.Payload,
		[]string{},
		player,
		update,
	) {
		return
	}

	// Verifico se vuole tornare indietro di stato
	if !proxy {
		if c.BackTo(1, &ShipController{}) {
			return
		}
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := services.NnSDK.UpdatePlayerState(ctx, &pb.UpdatePlayerStateRequest{
		PlayerState: c.State,
	})
	if err != nil {
		panic(err)
	}
	c.State = response.GetPlayerState()

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
	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		),
	)

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
						helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
					),
				),
			)

			return true, err
		}

		return false, err
	case 2:
		finishAt, err := ptypes.Timestamp(c.State.FinishAt)
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
			return false, err
		}
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *ShipRepairsController) Stage() (err error) {
	switch c.State.Stage {

	// In questo riporto al player le risorse e tempistiche necessarie alla riparazione della nave
	case 0:

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		response, err := services.NnSDK.GetPlayerShips(ctx, &pb.GetPlayerShipsRequest{
			PlayerID: c.Player.GetID(),
			Equipped: true,
		})
		if err != nil {
			return err
		}

		// Recupero nave player equipaggiata
		var playerShips []*pb.Ship
		playerShips = response.GetShips()

		// TODO: verificare, dovrebbe recuperarne solo una
		// Recupero name del player
		var playerShip *pb.Ship
		playerShip = playerShips[0]

		// Recupero informazioni nave da riparare
		responseRepairInfo, err := services.NnSDK.GetShipRepairInfo(ctx, &pb.GetShipRepairInfoRequest{
			Ship: playerShip,
		})
		if err != nil {
			return err
		}

		// Verifico se effettivamente la nave è da riparare
		var shipRecap string
		shipRecap = helpers.Trans(c.Player.Language.Slug, "ship.repairs.info")
		if responseRepairInfo.GetNeedRepairs() {
			shipRecap += fmt.Sprintf("🔧 %v/100%% (%s)\n%s\n%s ",
				playerShip.ShipStats.Integrity, helpers.Trans(c.Player.Language.Slug, "integrity"),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.time", responseRepairInfo.GetRepairTime()),
				helpers.Trans(c.Player.Language.Slug, "ship.repairs.quantity_resources", responseRepairInfo.GetQuantityResources(), responseRepairInfo.GetTypeResources()),
			)
		} else {
			shipRecap = helpers.Trans(c.Player.Language.Slug, "ship.repairs.dont_need")
		}

		// Aggiongo bottone start riparazione
		var keyboardRow [][]tgbotapi.KeyboardButton
		if responseRepairInfo.GetNeedRepairs() {
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
		c.Payload.Ship = playerShip
		c.Payload.QuantityResources = responseRepairInfo.GetQuantityResources()
		c.Payload.RepairTime = responseRepairInfo.GetRepairTime()
		c.Payload.TypeResources = responseRepairInfo.GetTypeResources()
		c.State.Stage = 1

	// In questo stage avvio effettivamente la riparzione
	case 1:

		// Avvio riparazione nave
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		response, err := services.NnSDK.StartShipRepair(ctx, &pb.StartShipRepairRequest{
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
		for _, resourceUsed := range response.GetStartShipRepair() {

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			response, err := services.NnSDK.GetResourceByID(ctx, &pb.GetResourceByIDRequest{
				ID: resourceUsed.ResourceID,
			})
			if err != nil {
				return err
			}

			var resource *pb.Resource
			resource = response.GetResource()

			recapResourceUsed += fmt.Sprintf("\n- %s x %v", resource.Name, resourceUsed.Quantity)
		}

		// Setto timer recuperato dalla chiamata delle info
		finishTime := helpers.GetEndTime(0, int(c.Payload.RepairTime), 0)
		c.State.FinishAt, _ = ptypes.TimestampProto(finishTime)

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
		c.State.ToNotify = true
		c.State.Stage = 2
		c.Breaker.ToMenu = true
	case 2:
		// Fine riparazione
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := services.NnSDK.EndShipRepair(ctx, &pb.EndShipRepairRequest{
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
		c.State.Completed = true
	}

	return
}
