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
			PlanetType: []string{"safe"},
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

	// Messaggi
	introMessage := helpers.Trans(player.Language.Slug, "safeplanet.coalition.research.info")

	var recapMessage string
	if rGetRecapActiveResearch.GetMissingResourcesCounter() > 0 {
		recapMessage += helpers.Trans(player.Language.Slug, "safeplanet.coalition.research.recap",
			rGetRecapActiveResearch.GetMissingResourcesCounter(),
			rGetRecapActiveResearch.GetResearch().GetRarity().GetSlug(),
		)
	} else {
		recapMessage += helpers.Trans(player.Language.Slug, "safeplanet.coalition.research.done")
	}

	// Keyboard
	var keyboardRow [][]tgbotapi.KeyboardButton
	if rGetRecapActiveResearch.GetMissingResourcesCounter() > 0 {
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.research.donation")),
		))
	}

	// Aggiungo tasti back and clears
	keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
	))

	msg := helpers.NewMessage(c.Update.Message.Chat.ID, fmt.Sprintf("%s\n\n%s", introMessage, recapMessage))
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard:       keyboardRow,
	}

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
