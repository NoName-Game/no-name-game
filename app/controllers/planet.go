package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
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
func (c *PlanetController) Handle(player nnsdk.Player, update tgbotapi.Update, proxy bool) {
	var err error
	c.Player = player
	c.Update = update
	c.Controller = "route.planet"

	// Se tutto ok imposto e setto il nuovo stato su redis
	_ = helpers.SetRedisState(c.Player, c.Controller)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &MenuController{}) {
			return
		}
	}

	planet, err := helpers.GetPlayerCurrentPlanet(c.Player)
	if err != nil {
		panic(err)
	}

	planetDetailsMsg := fmt.Sprintf("%s \n\n%s \n\n%s \n%s \n%s \n\n%s \n%s \n%s",
		helpers.Trans(player.Language.Slug, "planet.intro"),
		helpers.Trans(player.Language.Slug, "planet.details.system", planet.PlanetSystem.Name),
		helpers.Trans(player.Language.Slug, "planet.details.name", planet.Name),
		helpers.Trans(player.Language.Slug, "planet.details.biome", planet.Biome.Name),
		helpers.Trans(player.Language.Slug, "planet.details.atmosphere", planet.Atmosphere.Name),

		helpers.Trans(player.Language.Slug, "planet.details.coordinate.x", planet.X),
		helpers.Trans(player.Language.Slug, "planet.details.coordinate.y", planet.Y),
		helpers.Trans(player.Language.Slug, "planet.details.coordinate.z", planet.Z),
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
