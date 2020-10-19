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

	// Riceco pianeta per ID, in modo da ottenere maggior informazioni
	var rGetPlanetByID *pb.GetPlanetByIDResponse
	if rGetPlanetByID, err = config.App.Server.Connection.GetPlanetByID(helpers.NewContext(1), &pb.GetPlanetByIDRequest{
		PlanetID: playerPosition.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Aggiunto informazioni principali pianeta
	planetDetailsMsg := fmt.Sprintf("%s \n\n%s \n%s\n\n",
		helpers.Trans(player.Language.Slug, "planet.intro"),
		helpers.Trans(player.Language.Slug, "planet.details.system", rGetPlanetByID.GetPlanet().GetPlanetSystem().GetName()),
		helpers.Trans(player.Language.Slug, "planet.details.name", rGetPlanetByID.GetPlanet().GetName()),
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

	// Aggiungo coordinate pianeta
	planetDetailsMsg += fmt.Sprintf("📡 %s:\n%s \n%s \n%s",
		helpers.Trans(player.Language.Slug, "coordinate"),
		helpers.Trans(player.Language.Slug, "planet.details.coordinate.x", rGetPlanetByID.GetPlanet().GetX()),
		helpers.Trans(player.Language.Slug, "planet.details.coordinate.y", rGetPlanetByID.GetPlanet().GetY()),
		helpers.Trans(player.Language.Slug, "planet.details.coordinate.z", rGetPlanetByID.GetPlanet().GetZ()),
	)

	msg := helpers.NewMessage(c.Update.Message.Chat.ID, planetDetailsMsg)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
		),
	)

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
