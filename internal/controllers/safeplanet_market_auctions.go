package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Market
// ====================================
type SafePlanetMarketAuctionsController struct {
	Controller
}

func (c *SafePlanetMarketAuctionsController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.market.auctions",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetMarketController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
			AllowedControllers: []string{
				"route.safeplanet.market.auctions.buy",
				"route.safeplanet.market.auctions.sell",
				"route.safeplanet.market.auctions.my_auction",
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMarketAuctionsController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	msg := helpers.NewMessage(c.ChatID,
		fmt.Sprintf("%s\n\n%s",
			helpers.Trans(player.Language.Slug, "safeplanet.market.auctions.info"),
			helpers.Trans(player.Language.Slug, "safeplanet.market.auctions.tip"),
		),
	)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.market.auctions.buy")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.market.auctions.sell")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.market.auctions.my_auction")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *SafePlanetMarketAuctionsController) Validator() bool {
	return false
}

func (c *SafePlanetMarketAuctionsController) Stage() {
	//
}
