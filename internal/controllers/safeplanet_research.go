package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetResearchController
// ====================================
type SafePlanetResearchController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetResearchController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.research",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 1,
			},
		},
	}) {
		return
	}

	// Recupero informazioni ricerca
	var err error
	var rGetRecapActiveResearch *pb.GetRecapActiveResearchResponse
	if rGetRecapActiveResearch, err = config.App.Server.Connection.GetRecapActiveResearch(helpers.NewContext(1), &pb.GetRecapActiveResearchRequest{}); err != nil {
		c.Logger.Panic(err)
	}

	message := fmt.Sprintf("%s\n\n%s",
		helpers.Trans(player.Language.Slug, "safeplanet.coalition.research.info"),
		helpers.Trans(player.Language.Slug, "safeplanet.coalition.research.recap",
			rGetRecapActiveResearch.GetMissingResourcesCounter(),
			rGetRecapActiveResearch.GetResearch().GetRarity().GetName(),
		),
	)

	msg := helpers.NewMessage(c.Update.Message.Chat.ID, message)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.research.donation")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
		),
	)

	if _, err := helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}
func (c *SafePlanetResearchController) Validator() bool {
	return false
}

func (c *SafePlanetResearchController) Stage() {
	//
}
