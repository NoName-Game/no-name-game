package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Coalition
// ====================================
type SafePlanetCoalitionStatisticsController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetCoalitionStatisticsController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.statistics",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
		},
	}) {
		return
	}

	switch c.Update.Message.Text {
	case helpers.Trans(player.Language.Slug, "safeplanet.coalition.statistics.top_kill_enemy"):
		c.TopKillEnemy()
	case helpers.Trans(player.Language.Slug, "safeplanet.coalition.statistics.top_explored_planets"):
		c.TopExploredPlanets()
	case helpers.Trans(player.Language.Slug, "safeplanet.coalition.statistics.top_travel"):
		c.TopTravel()
	case helpers.Trans(player.Language.Slug, "safeplanet.coalition.statistics.top_mission_completed"):
		c.TopMissionCompleted()
	}

	msg := helpers.NewMessage(c.ChatID, helpers.Trans(player.Language.Slug, "safeplanet.coalition.statistics.intro"))
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "safeplanet.coalition.statistics.top_kill_enemy")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "safeplanet.coalition.statistics.top_mission_completed")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "safeplanet.coalition.statistics.top_travel")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "safeplanet.coalition.statistics.top_explored_planets")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *SafePlanetCoalitionStatisticsController) Validator() bool {
	return false
}

func (c *SafePlanetCoalitionStatisticsController) Stage() {
	//
}

// TopMission
func (c *SafePlanetCoalitionStatisticsController) TopMissionCompleted() {
	var err error

	var rStatisticsTopMissionAll *pb.StatisticsTopMissionAllResponse
	if rStatisticsTopMissionAll, err = config.App.Server.Connection.StatisticsTopMissionAll(helpers.NewContext(1), &pb.StatisticsTopMissionAllRequest{}); err != nil {
		c.Logger.Panic(err)
	}

	// Preparo lista ALL
	var recapAll string
	for i, result := range rStatisticsTopMissionAll.GetResults() {
		recapAll += c.AllFormatter(i+1, result.GetUsername(), result.GetResult())
	}

	var rStatisticsTopMissionYou *pb.StatisticsTopMissionYouResponse
	if rStatisticsTopMissionYou, err = config.App.Server.Connection.StatisticsTopMissionYou(helpers.NewContext(1), &pb.StatisticsTopMissionYouRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	recapYou := c.YouFormatter(rStatisticsTopMissionYou.GetResult().GetUsername(), rStatisticsTopMissionYou.GetResult().Result)

	// Send
	c.SendStatistcs("top_mission_completed", recapAll, recapYou)
}

// TopTravel
func (c *SafePlanetCoalitionStatisticsController) TopTravel() {
	var err error

	var rStatisticsTopTravelAll *pb.StatisticsTopTravelAllResponse
	if rStatisticsTopTravelAll, err = config.App.Server.Connection.StatisticsTopTravelAll(helpers.NewContext(1), &pb.StatisticsTopTravelAllRequest{}); err != nil {
		c.Logger.Panic(err)
	}

	// Preparo lista ALL
	var recapAll string
	for i, result := range rStatisticsTopTravelAll.GetResults() {
		recapAll += c.AllFormatter(i+1, result.GetUsername(), result.GetResult())
	}

	var rStatisticsTopTravelYou *pb.StatisticsTopTravelYouResponse
	if rStatisticsTopTravelYou, err = config.App.Server.Connection.StatisticsTopTravelYou(helpers.NewContext(1), &pb.StatisticsTopTravelYouRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	recapYou := c.YouFormatter(rStatisticsTopTravelYou.GetResult().GetUsername(), rStatisticsTopTravelYou.GetResult().Result)

	// Send
	c.SendStatistcs("top_travel", recapAll, recapYou)
}

// TopExploredPlanets
func (c *SafePlanetCoalitionStatisticsController) TopExploredPlanets() {
	var err error

	var rStatisticsTopPlanetExploredAll *pb.StatisticsTopPlanetExploredAllResponse
	if rStatisticsTopPlanetExploredAll, err = config.App.Server.Connection.StatisticsTopPlanetExploredAll(helpers.NewContext(1), &pb.StatisticsTopPlanetExploredAllRequest{}); err != nil {
		c.Logger.Panic(err)
	}

	// Preparo lista ALL
	var recapAll string
	for i, result := range rStatisticsTopPlanetExploredAll.GetResults() {
		recapAll += c.AllFormatter(i+1, result.GetUsername(), result.GetResult())
	}

	var rStatisticsTopPlanetExploredYou *pb.StatisticsTopPlanetExploredYouResponse
	if rStatisticsTopPlanetExploredYou, err = config.App.Server.Connection.StatisticsTopPlanetExploredYou(helpers.NewContext(1), &pb.StatisticsTopPlanetExploredYouRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	recapYou := c.YouFormatter(rStatisticsTopPlanetExploredYou.GetResult().GetUsername(), rStatisticsTopPlanetExploredYou.GetResult().Result)

	// Send
	c.SendStatistcs("top_explored_planets", recapAll, recapYou)
}

// TopkillEnemy
func (c *SafePlanetCoalitionStatisticsController) TopKillEnemy() {
	var err error

	var rStatisticsTopPlayerEnemyKillAll *pb.StatisticsTopPlayerEnemyKillAllResponse
	if rStatisticsTopPlayerEnemyKillAll, err = config.App.Server.Connection.StatisticsTopPlayerEnemyKillAll(helpers.NewContext(1), &pb.StatisticsTopPlayerEnemyKillAllRequest{}); err != nil {
		c.Logger.Panic(err)
	}

	// Preparo lista ALL
	var recapAll string
	for i, result := range rStatisticsTopPlayerEnemyKillAll.GetResults() {
		recapAll += c.AllFormatter(i+1, result.GetUsername(), result.GetResult())
	}

	var rStatisticsTopPlayerEnemyKillYou *pb.StatisticsTopPlayerEnemyKillYouResponse
	if rStatisticsTopPlayerEnemyKillYou, err = config.App.Server.Connection.StatisticsTopPlayerEnemyKillYou(helpers.NewContext(1), &pb.StatisticsTopPlayerEnemyKillYouRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	recapYou := c.YouFormatter(rStatisticsTopPlayerEnemyKillYou.GetResult().GetUsername(), rStatisticsTopPlayerEnemyKillYou.GetResult().Result)

	// Send
	c.SendStatistcs("top_kill_enemy", recapAll, recapYou)
}

func (c *SafePlanetCoalitionStatisticsController) AllFormatter(index int, username string, result int64) string {
	return fmt.Sprintf("%v) %s *%v*\n", index, username, result)
}

func (c *SafePlanetCoalitionStatisticsController) YouFormatter(username string, result int64) string {
	return fmt.Sprintf("*%s:* *%v*\n", username, result)
}

func (c *SafePlanetCoalitionStatisticsController) SendStatistcs(title, all, you string) {
	var err error
	msg := helpers.NewMessage(c.ChatID, fmt.Sprintf("*%s*\n%s\n\n%s\n%s",
		helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("safeplanet.coalition.statistics.%s", title)),
		helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("safeplanet.coalition.statistics.%s.description", title)),
		all,
		you,
	))
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}
