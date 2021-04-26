package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/config"
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipController
// Ogni player ha la possibilità di spostarsi nei diversi pianeti
// del sistema di NoName
// ====================================
type ShipController struct {
	Controller
}

func (c *ShipController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship",
		},
		Configurations: ControllerConfigurations{
			ControllerBlocked: []string{"exploration", "hunting"},
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
			AllowedControllers: []string{
				"route.assault",
				"route.ship.rests",
				"route.ship.travel",
				"route.ship.laboratory",
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// ****************************
	// Recupero nave attiva de player
	// ****************************
	var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
	if rGetPlayerShipEquipped, err = config.App.Server.Connection.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Invio messaggio
	msg := helpers.NewMessage(c.ChatID,
		fmt.Sprintf("%s %s %s %s %s %s %s",
			helpers.Trans(c.Player.Language.Slug, "ship.intro"),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_stats", rGetPlayerShipEquipped.GetShip().GetName(), rGetPlayerShipEquipped.GetShip().GetRarity().GetSlug()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_engine", rGetPlayerShipEquipped.GetShip().GetEngine()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_type", helpers.Trans(c.Player.GetLanguage().GetSlug(), fmt.Sprintf("ship.category.%s", rGetPlayerShipEquipped.GetShip().GetShipCategory().GetSlug()))),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_scanner", rGetPlayerShipEquipped.GetShip().GetRadar()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_integrity", rGetPlayerShipEquipped.GetShip().GetIntegrity()),
			helpers.Trans(c.Player.Language.Slug, "ship.travel.ship_carburante", rGetPlayerShipEquipped.GetShip().GetTank()),
		))
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.travel")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.rests")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.laboratory")),
		),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.assault"))),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		),
	)

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *ShipController) Validator() bool {
	return false
}

func (c *ShipController) Stage() {
	//
}
