package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Domain
// ====================================
type PlanetDomainController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlanetDomainController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.planet.domain",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlanetController{},
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

	// Recupero top 10 gilde per uccisioni in questo pianeta
	var rGetDomainsByPlanetID *pb.GetDomainsByPlanetIDResponse
	if rGetDomainsByPlanetID, err = config.App.Server.Connection.GetDomainsByPlanetID(helpers.NewContext(1), &pb.GetDomainsByPlanetIDRequest{
		PlanetID: playerPosition.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Intro msg
	var domainsListMsg string
	domainsListMsg = helpers.Trans(player.Language.Slug, "domain.intro")

	// Eseguo recap conquistatori
	domainsListMsg += helpers.Trans(player.Language.Slug, "domain.list.intro")
	for i, domains := range rGetDomainsByPlanetID.GetDomains() {
		if i < 1 {
			domainsListMsg += fmt.Sprintf("🚩 💢 <b>%s</b> ⚔️ <b>%d</b> \n",
				domains.GetGuild().GetName(),
				domains.GetNKills(),
			)
			continue
		}

		domainsListMsg += fmt.Sprintf("%d - 💢 %s ⚔️ %d\n",
			i+1,
			domains.GetGuild().GetName(),
			domains.GetNKills(),
		)
	}

	// Nessun conquistatore
	if len(rGetDomainsByPlanetID.GetDomains()) < 1 {
		domainsListMsg += helpers.Trans(player.Language.Slug, "domain.planet_free")
	}

	msg := helpers.NewMessage(c.ChatID, domainsListMsg)
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

func (c *PlanetDomainController) Validator() bool {
	return false
}

func (c *PlanetDomainController) Stage() {
	//
}
