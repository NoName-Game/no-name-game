package controllers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"
)

// ====================================
// SafePlanetHangarController
// ====================================
type SafePlanetHangarController struct {
	Controller
}

func (c *SafePlanetHangarController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.hangar",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
			AllowedControllers: []string{
				"route.safeplanet.hangar.ships",
				"route.safeplanet.hangar.repair",
				"route.safeplanet.hangar.create",
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetHangarController) Handle(player *pb.Player, update tgbotapi.Update) {
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
	recapShip := helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.ship_recap",
		rGetPlayerShipEquipped.GetShip().GetName(), strings.ToUpper(rGetPlayerShipEquipped.GetShip().GetRarity().GetSlug()),
		rGetPlayerShipEquipped.GetShip().GetShipCategory().GetName(),
		rGetPlayerShipEquipped.GetShip().GetIntegrity(), helpers.Trans(c.Player.Language.Slug, "integrity"),
		rGetPlayerShipEquipped.GetShip().GetTank(), helpers.Trans(c.Player.Language.Slug, "fuel"),
	)

	msg := helpers.NewMessage(c.ChatID, fmt.Sprintf("%s\n\n%s",
		helpers.Trans(player.Language.Slug, "safeplanet.hangar.intro"),
		recapShip,
	))
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.hangar.ships")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.hangar.repair")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.hangar.create")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *SafePlanetHangarController) Validator() bool {
	return false
}

func (c *SafePlanetHangarController) Stage() {
	//
}
