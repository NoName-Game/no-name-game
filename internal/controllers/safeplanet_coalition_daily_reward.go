package controllers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/grpc/status"
	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"
)

// ====================================
// SafePlanetCoalitionDailyRewardController
// ====================================
type SafePlanetCoalitionDailyRewardController struct {
	Controller
}

func (c *SafePlanetCoalitionDailyRewardController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.daily_reward",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetCoalitionDailyRewardController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Recupero informazioni daily reward
	var err error
	var rGetPlayerDailyReward *pb.GetPlayerDailyRewardResponse
	rGetPlayerDailyReward, err = config.App.Server.Connection.GetPlayerDailyReward(helpers.NewContext(1), &pb.GetPlayerDailyRewardRequest{
		PlayerID: c.Player.ID,
	})

	var dailyRewardMessage string
	if err != nil {
		if errDetails, _ := status.FromError(err); errDetails.Message() == "daily reward already taken" {
			dailyRewardMessage = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.daily_reward_already_taken")
		} else {
			c.Logger.Panic(err)
		}
	} else {
		// Recupero dettaglio Item droppato
		itemFound := helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("items.%s", rGetPlayerDailyReward.GetItem().GetSlug()))

		dailyRewardMessage = helpers.Trans(c.Player.Language.Slug, "safeplanet.coalition.daily_reward",
			rGetPlayerDailyReward.GetExperience(),
			rGetPlayerDailyReward.GetMoney(),
			rGetPlayerDailyReward.GetDiamond(),
			itemFound,
		)
	}

	msg := helpers.NewMessage(c.ChatID, dailyRewardMessage)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}

	// Completo progressione
	c.CurrentState.Completed = true
	c.Completing(nil)
}

func (c *SafePlanetCoalitionDailyRewardController) Validator() bool {
	return false
}

func (c *SafePlanetCoalitionDailyRewardController) Stage() {
	//
}
