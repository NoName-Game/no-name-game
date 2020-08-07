package controllers

import (
	"fmt"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

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
func (c *PlanetController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	var err error
	c.Player = player
	c.Update = update
	c.Controller = "route.planet"

	// Se tutto ok imposto e setto il nuovo stato su redis
	_ = helpers.SetRedisState(*c.Player, c.Controller)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &MenuController{}) {
			return
		}
	}

	var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
	rGetPlayerCurrentPlanet, err = services.NnSDK.GetPlayerCurrentPlanet(helpers.NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}

	planetDetailsMsg := fmt.Sprintf("%s \n\n%s \n\n%s \n%s \n%s \n\n%s \n%s \n%s",
		helpers.Trans(player.Language.Slug, "planet.intro"),
		// helpers.Trans(player.Language.Slug, "planet.details.system", planet.PlanetSystem.Name),
		helpers.Trans(player.Language.Slug, "planet.details.name", rGetPlayerCurrentPlanet.GetPlanet().GetName()),
		helpers.Trans(player.Language.Slug, "planet.details.biome", rGetPlayerCurrentPlanet.GetPlanet().GetBiome().GetName()),
		helpers.Trans(player.Language.Slug, "planet.details.atmosphere", rGetPlayerCurrentPlanet.GetPlanet().GetAtmosphere().GetName()),

		helpers.Trans(player.Language.Slug, "planet.details.coordinate.x", rGetPlayerCurrentPlanet.GetPlanet().GetX()),
		helpers.Trans(player.Language.Slug, "planet.details.coordinate.y", rGetPlayerCurrentPlanet.GetPlanet().GetY()),
		helpers.Trans(player.Language.Slug, "planet.details.coordinate.z", rGetPlayerCurrentPlanet.GetPlanet().GetZ()),
	)

	msg := services.NewMessage(c.Update.Message.Chat.ID, planetDetailsMsg)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.back")),
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
