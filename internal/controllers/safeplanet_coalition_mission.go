package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetMissionController
// ====================================
type SafePlanetMissionController struct {
	Controller
}

func (c *SafePlanetMissionController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.mission",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 1,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
			AllowedControllers: []string{
				"route.safeplanet.coalition.mission.search",
				"route.safeplanet.coalition.mission.complete",
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMissionController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	switch c.Update.Message.Text {
	case helpers.Trans(player.Language.Slug, "safeplanet.mission.statistic.top"):
		c.TopMission()
		return
	case helpers.Trans(player.Language.Slug, "safeplanet.mission.statistic.bad"):
		c.BadMission()
		return
	}

	// Keyboard
	var keyboardRow [][]tgbotapi.KeyboardButton

	var rCheckMission *pb.CheckMissionResponse
	if rCheckMission, err = config.App.Server.Connection.CheckMission(helpers.NewContext(1), &pb.CheckMissionRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Verifico se il player deve completare o non una missione
	if rCheckMission.GetInMission() {
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.coalition.mission.complete")),
		))
	} else {
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.safeplanet.coalition.mission.search")),
		))
	}

	// Aggiungo tasti top and back
	keyboardRow = append(keyboardRow,
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "safeplanet.mission.statistic.top")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "safeplanet.mission.statistic.bad")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)

	msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "safeplanet.mission.info"))
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard:       keyboardRow,
	}

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *SafePlanetMissionController) Validator() bool {
	return false
}

func (c *SafePlanetMissionController) Stage() {
	//
}

func (c *SafePlanetMissionController) TopMission() {
	var err error
	var rGetMissionsMostCompleted *pb.GetMissionsMostCompletedResponse
	if rGetMissionsMostCompleted, err = config.App.Server.Connection.GetMissionsMostCompleted(helpers.NewContext(1), &pb.GetMissionsMostCompletedRequest{}); err != nil {
		c.Logger.Panic(err)
	}

	var topMissions string
	for _, mission := range rGetMissionsMostCompleted.GetMissions() {
		// Recupero quotazione della missione allo stato attuale
		var rCheckMissionReward *pb.CheckMissionRewardResponse
		if rCheckMissionReward, err = config.App.Server.Connection.CheckMissionReward(helpers.NewContext(1), &pb.CheckMissionRewardRequest{
			MissionID: mission.GetMission().GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		rewards := fmt.Sprintf("+<b>%v</b>💰 +<b>%v</b>🏵",
			rCheckMissionReward.GetMoney(),
			rCheckMissionReward.GetExp(),
		)
		if rCheckMissionReward.GetDiamond() > 0 {
			rewards += fmt.Sprintf(" +<b>%v</b>💎", rCheckMissionReward.GetDiamond())
		}

		topMissions += fmt.Sprintf("✅ %v | #M%v | %s\n", mission.GetCounter(), mission.GetMission().GetID(), rewards)
	}

	msg := helpers.NewMessage(c.ChatID, topMissions)
	msg.ParseMode = tgbotapi.ModeHTML

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *SafePlanetMissionController) BadMission() {
	var err error
	var rGetMissionsMostLeaved *pb.GetMissionsMostLeavedResponse
	if rGetMissionsMostLeaved, err = config.App.Server.Connection.GetMissionsMostLeaved(helpers.NewContext(1), &pb.GetMissionsMostLeavedRequest{}); err != nil {
		c.Logger.Panic(err)
	}

	var topMissions string
	for _, mission := range rGetMissionsMostLeaved.GetMissions() {
		// Recupero quotazione della missione allo stato attuale
		var rCheckMissionReward *pb.CheckMissionRewardResponse
		if rCheckMissionReward, err = config.App.Server.Connection.CheckMissionReward(helpers.NewContext(1), &pb.CheckMissionRewardRequest{
			MissionID: mission.GetMission().GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		rewards := fmt.Sprintf("+<b>%v</b>💰 +<b>%v</b>🏵",
			rCheckMissionReward.GetMoney(),
			rCheckMissionReward.GetExp(),
		)
		if rCheckMissionReward.GetDiamond() > 0 {
			rewards += fmt.Sprintf(" +<b>%v</b>💎", rCheckMissionReward.GetDiamond())
		}

		topMissions += fmt.Sprintf("❌ %v | #M%v | %s\n", mission.GetCounter(), mission.GetMission().GetID(), rewards)
	}

	msg := helpers.NewMessage(c.ChatID, topMissions)
	msg.ParseMode = tgbotapi.ModeHTML

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}
