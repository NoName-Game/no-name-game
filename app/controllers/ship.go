package controllers

import (
	"strings"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
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

// ====================================
// Handle
// ====================================
func (c *ShipController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error

	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
		},
	}) {
		return
	}

	// ****************************
	// Recupero nave attiva de player
	// ****************************
	var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
	if rGetPlayerShipEquipped, err = services.NnSDK.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		panic(err)
	}

	// Invio messaggio
	msg := services.NewMessage(c.Update.Message.Chat.ID,
		helpers.Trans(c.Player.Language.Slug, "ship.report.card",
			rGetPlayerShipEquipped.GetShip().Name, strings.ToUpper(rGetPlayerShipEquipped.GetShip().GetRarity().GetSlug()),
			rGetPlayerShipEquipped.GetShip().GetShipCategory().GetName(),
			rGetPlayerShipEquipped.GetShip().GetShipStats().GetIntegrity(), helpers.Trans(c.Player.Language.Slug, "integrity"),
			rGetPlayerShipEquipped.GetShip().GetShipStats().GetTank(), helpers.Trans(c.Player.Language.Slug, "fuel"),
		),
	)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.travel")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.rests")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.repairs")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.ship.laboratory")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		),
	)

	if _, err = services.SendMessage(msg); err != nil {
		panic(err)
	}
}

func (c *ShipController) Validator() {
	//
}

func (c *ShipController) Stage() {
	//
}
