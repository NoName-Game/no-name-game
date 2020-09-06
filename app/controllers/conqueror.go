package controllers

import (
	"fmt"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Conqueror
// ====================================
type ConquerorController struct {
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *ConquerorController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error
	c.Player = player
	c.Update = update
	c.Configuration.Controller = "route.conqueror"

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

	// Recupero top 10 player per uccisioni in questo pianeta
	var rGetConquerorsByPlanetID *pb.GetConquerorsByPlanetIDResponse
	rGetConquerorsByPlanetID, err = services.NnSDK.GetConquerorsByPlanetID(helpers.NewContext(1), &pb.GetConquerorsByPlanetIDRequest{
		PlanetID: rGetPlayerCurrentPlanet.GetPlanet().GetID(),
	})
	if err != nil {
		panic(err)
	}

	// Intro msg
	var conquerorsListMsg string
	conquerorsListMsg = helpers.Trans(player.Language.Slug, "conqueror.intro")

	// Eseguo recap conquistatori
	conquerorsListMsg += helpers.Trans(player.Language.Slug, "conqueror.list.intro")
	for i, conquerors := range rGetConquerorsByPlanetID.GetConquerors() {
		if i < 1 {
			conquerorsListMsg += fmt.Sprintf("\n- 👨🏼‍🚀 *%s* ⚔️ *%d* 🚩",
				conquerors.GetPlayer().GetUsername(),
				conquerors.GetNKills(),
			)
			continue
		}

		conquerorsListMsg += fmt.Sprintf("\n- 👨🏼‍🚀 %s ⚔️ %d",
			conquerors.GetPlayer().GetUsername(),
			conquerors.GetNKills(),
		)
	}

	if len(rGetConquerorsByPlanetID.GetConquerors()) < 1 {
		conquerorsListMsg += "Non esiste nessun player che ha conquistato un cazzo"
	}

	msg := services.NewMessage(c.Update.Message.Chat.ID, conquerorsListMsg)
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

func (c *ConquerorController) Validator() {
	//
}

func (c *ConquerorController) Stage() {
	//
}
