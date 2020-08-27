package controllers

import (
	"fmt"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Planet
// ====================================
type PlanetController struct {
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *PlanetController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error
	c.Player = player
	c.Update = update
	c.Configuration.Controller = "route.planet"

	// Se tutto ok imposto e setto il nuovo stato in cache
	helpers.SetCacheState(c.Player.ID, c.Configuration.Controller)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if c.BackTo(0, &MenuController{}) {
		return
	}

	// Recupero pianeta corrente del player
	var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
	rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}

	// Riceco pianeta per ID, in modo da ottenere maggior informazioni
	var rGetPlanetByID *pb.GetPlanetByIDResponse
	rGetPlanetByID, err = services.NnSDK.GetPlanetByID(helpers.NewContext(1), &pb.GetPlanetByIDRequest{
		PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
	})
	if err != nil {
		panic(err)
	}

	// Aggiunto informazioni principali pianeta
	planetDetailsMsg := fmt.Sprintf("%s \n\n%s \n%s \n%s \n%s \n\n",
		helpers.Trans(player.Language.Slug, "planet.intro"),
		helpers.Trans(player.Language.Slug, "planet.details.system", rGetPlanetByID.GetPlanet().GetPlanetSystem().GetName()),
		helpers.Trans(player.Language.Slug, "planet.details.name", rGetPlanetByID.GetPlanet().GetName()),
		helpers.Trans(player.Language.Slug, "planet.details.biome", rGetPlanetByID.GetPlanet().GetBiome().GetName()),
		helpers.Trans(player.Language.Slug, "planet.details.atmosphere", rGetPlanetByID.GetPlanet().GetAtmosphere().GetName()),
	)

	var rCountPlayerVisitedCurrentPlanet *pb.CountPlayerVisitedCurrentPlanetResponse
	rCountPlayerVisitedCurrentPlanet, err = services.NnSDK.CountPlayerVisitedCurrentPlanet(helpers.NewContext(1), &pb.CountPlayerVisitedCurrentPlanetRequest{
		PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
	})
	if err != nil {
		panic(err)
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

	msg := services.NewMessage(c.Update.Message.Chat.ID, planetDetailsMsg)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
		),
	)

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}

func (c *PlanetController) Validator() {
	//
}

func (c *PlanetController) Stage() {
	//
}
