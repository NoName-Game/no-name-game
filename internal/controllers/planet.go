package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Planet
// ====================================
type PlanetController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlanetController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.planet",
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

	// Recupero posizione player corrente
	var playerPosition *pb.Planet
	if playerPosition, err = helpers.GetPlayerPosition(c.Player.ID); err != nil {
		c.Logger.Panic(err)
	}

	// Ricerco pianeta per ID, in modo da ottenere maggior informazioni
	var rGetPlanetByID *pb.GetPlanetByIDResponse
	if rGetPlanetByID, err = config.App.Server.Connection.GetPlanetByID(helpers.NewContext(1), &pb.GetPlanetByIDRequest{
		PlanetID: playerPosition.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Aggiunto informazioni principali pianeta
	planetDetailsMsg := fmt.Sprintf("%s\n%s \n%s\n%s\n\n",
		helpers.Trans(player.Language.Slug, "planet.intro"),
		helpers.Trans(player.Language.Slug, "planet.details.system", rGetPlanetByID.GetPlanet().GetPlanetSystem().GetName()),
		helpers.Trans(player.Language.Slug, "planet.details.name", rGetPlanetByID.GetPlanet().GetName()),
		helpers.Trans(player.Language.Slug, "planet.details.coordinate.encypted", rGetPlanetByID.GetPlanet().GetHashPosition()),
	)

	var rCountPlayerVisitedCurrentPlanet *pb.CountPlayerVisitedCurrentPlanetResponse
	if rCountPlayerVisitedCurrentPlanet, err = config.App.Server.Connection.CountPlayerVisitedCurrentPlanet(helpers.NewContext(1), &pb.CountPlayerVisitedCurrentPlanetRequest{
		PlanetID: playerPosition.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Aggiunto informazioni aggiuntive
	planetDetailsMsg += fmt.Sprintf("%s\n\n",
		helpers.Trans(player.Language.Slug, "planet.details.count_visited_player", rCountPlayerVisitedCurrentPlanet.GetValue()),
	)

	// Verifico se il player ha già il pianeta tra i preferiti
	var rCheckIfPlayerHavePlanetBookmark *pb.CheckIfPlayerHavePlanetBookmarkResponse
	if rCheckIfPlayerHavePlanetBookmark, err = config.App.Server.Connection.CheckIfPlayerHavePlanetBookmark(helpers.NewContext(1), &pb.CheckIfPlayerHavePlanetBookmarkRequest{
		PlanetID: playerPosition.GetID(),
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	var keyboardRow [][]tgbotapi.KeyboardButton
	if rCheckIfPlayerHavePlanetBookmark.GetHaveBookmark() {
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.planet.bookmark.remove")),
		))
	} else {
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.planet.bookmark.add")),
		))
	}

	// Aggiungo torna al menu
	keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
	))

	msg := helpers.NewMessage(c.Update.Message.Chat.ID, planetDetailsMsg)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard:       keyboardRow,
	}

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *PlanetController) Validator() bool {
	return false
}

func (c *PlanetController) Stage() {
	//
}
