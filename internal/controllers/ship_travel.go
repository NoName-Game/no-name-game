package controllers

import (
	"fmt"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipTravelController
// Ogni player ha la possibilità di spostarsi nei diversi pianeti
// del sistema di NoName
// ====================================
type ShipTravelController struct {
	Controller
}

func (c *ShipTravelController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship.travel",
		},
		Configurations: ControllerConfigurations{
			ControllerBlocked: []string{"exploration", "hunting"},
			ControllerBack: ControllerBack{
				To:        &ShipController{},
				FromStage: 0,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipTravelController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Recupero nave attualemente attiva
	var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
	if rGetPlayerShipEquipped, err = config.App.Server.Connection.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Sostituisco bottone navigazione con atterra
	var travel = "route.ship.travel.finding"
	for _, state := range c.Data.PlayerActiveStates {
		if state.Controller == "route.ship.travel.finding" && state.GetStage() > 1 {
			// Se il player sta già viaggiando
			var finishAt time.Time
			if finishAt, err = helpers.GetEndTime(state.GetFinishAt(), c.Player); err != nil {
				c.Logger.Panic(err)
			}

			if time.Now().After(finishAt) {
				travel = "ship.travel.land"
			}
		}
	}

	// Invio messaggio con recap
	msg := helpers.NewMessage(c.ChatID,
		fmt.Sprintf("%s %s %s %s %s %s %s",
			helpers.Trans(c.Player.Language.Slug, "ship.travel.info"),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_stats"),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_engine", rGetPlayerShipEquipped.GetShip().GetEngine()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_scanner", rGetPlayerShipEquipped.GetShip().GetRadar()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_integrity", rGetPlayerShipEquipped.GetShip().GetIntegrity()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_carburante", rGetPlayerShipEquipped.GetShip().GetTank()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.tip"),
		))
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, travel)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.travel.favorite")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.travel.manual")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.travel.rescue")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		),
	)

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *ShipTravelController) Validator() bool {
	return false
}

func (c *ShipTravelController) Stage() {
	//
}
