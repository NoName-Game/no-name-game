package controllers

import (
	"fmt"

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

// ====================================
// Handle
// ====================================
func (c *ShipTravelController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship.travel",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &ShipController{},
				FromStage: 0,
			},
		},
	}) {
		return
	}

	// Recupero nave attualemente attiva
	var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
	if rGetPlayerShipEquipped, err = config.App.Server.Connection.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Invio messaggio con recap
	msg := helpers.NewMessage(c.Update.Message.Chat.ID,
		fmt.Sprintf("%s %s %s %s %s %s %s",
			helpers.Trans(c.Player.Language.Slug, "ship.travel.info"),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_stats"),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_engine", rGetPlayerShipEquipped.GetShip().GetEngine()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_scanner", rGetPlayerShipEquipped.GetShip().GetRadar()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_integrity", rGetPlayerShipEquipped.GetShip().GetIntegrity()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_carburante", rGetPlayerShipEquipped.GetShip().GetTank()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.tip"),
		))
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.travel.finding")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.travel.favorite")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
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
