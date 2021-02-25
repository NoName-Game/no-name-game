package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Conqueror
// ====================================
type ConquerorController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *ConquerorController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.conqueror",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			PlanetType: []string{"default"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
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

	// Recupero top 10 player per uccisioni in questo pianeta
	var rGetConquerorsByPlanetID *pb.GetConquerorsByPlanetIDResponse
	if rGetConquerorsByPlanetID, err = config.App.Server.Connection.GetConquerorsByPlanetID(helpers.NewContext(1), &pb.GetConquerorsByPlanetIDRequest{
		PlanetID: playerPosition.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Intro msg
	var conquerorsListMsg string
	conquerorsListMsg = helpers.Trans(player.Language.Slug, "conqueror.intro")

	// Eseguo recap conquistatori
	conquerorsListMsg += helpers.Trans(player.Language.Slug, "conqueror.list.intro")
	for i, conquerors := range rGetConquerorsByPlanetID.GetConquerors() {
		if i < 1 {
			conquerorsListMsg += fmt.Sprintf("🚩 👨🏼‍🚀 <b>%s</b> ⚔️ <b>%d</b> \n",
				conquerors.GetPlayer().GetUsername(),
				conquerors.GetNKills(),
			)
			continue
		}

		conquerorsListMsg += fmt.Sprintf("%d - 👨🏼‍🚀 %s ⚔️ %d\n",
			i+1,
			conquerors.GetPlayer().GetUsername(),
			conquerors.GetNKills(),
		)
	}

	// Nessun conquistatore
	if len(rGetConquerorsByPlanetID.GetConquerors()) < 1 {
		conquerorsListMsg += helpers.Trans(player.Language.Slug, "conqueror.planet_free")
	}

	msg := helpers.NewMessage(c.ChatID, conquerorsListMsg)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *ConquerorController) Validator() bool {
	return false
}

func (c *ConquerorController) Stage() {
	//
}
