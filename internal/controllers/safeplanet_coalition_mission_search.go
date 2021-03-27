package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetMissionSearchController
// ====================================
type SafePlanetMissionSearchController struct {
	Controller
	Payload struct {
		MissionID uint32
	}
}

func (c *SafePlanetMissionSearchController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.mission.search",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetMissionController{},
				FromStage: 1,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu", "route.breaker.continue"},
				1: {"route.breaker.menu", "route.breaker.clears", "route.breaker.continue"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMissionSearchController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Validate
	if c.Validator() {
		c.Validate()
		return
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(&c.Payload)
}

func (c *SafePlanetMissionSearchController) Validator() bool {
	var err error
	var rCheckMission *pb.CheckMissionResponse
	if rCheckMission, err = config.App.Server.Connection.CheckMission(helpers.NewContext(1), &pb.CheckMissionRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	// Verifico sempre se il player ha finito la missione
	if rCheckMission.GetInMission() && rCheckMission.GetCompleted() {
		c.CurrentState.Stage = 2
		return false
	}

	if c.CurrentState.Stage > 0 &&
		helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.leave") != c.Update.Message.Text &&
		helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.accept") != c.Update.Message.Text {
		// Verifico se realmente il player è in missione
		if rCheckMission.GetInMission() && !rCheckMission.GetCompleted() {
			c.CurrentState.Stage = 0
			return false
		}
	}

	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico avvio missione
	// ##################################################################################################
	case 1:
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.accept") == c.Update.Message.Text ||
			helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.leave") == c.Update.Message.Text {
			c.CurrentState.Stage = 1
			return false
		}

		return true
	}

	return false
}

func (c *SafePlanetMissionSearchController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		var rCheckMission *pb.CheckMissionResponse
		if rCheckMission, err = config.App.Server.Connection.CheckMission(helpers.NewContext(1), &pb.CheckMissionRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Verifico se il player è già in missione
		if rCheckMission.GetInMission() && !rCheckMission.GetCompleted() {
			// Recupero dettagli missione
			missionRecap := c.getMissionRecap(rCheckMission.GetMission())

			msg := helpers.NewMessage(c.ChatID, missionRecap)
			msg.ParseMode = tgbotapi.ModeHTML
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.continue")),
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.leave")),
				),
			)

			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}
		} else {
			// Chiamo il ws e recupero il tipo di missione da effettuare
			// attraverso il tipo di missione costruisco il corpo del messaggio
			var rNewMission *pb.NewMissionResponse
			if rNewMission, err = config.App.Server.Connection.NewMission(helpers.NewContext(1), &pb.NewMissionRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Recupero dettagli missione
			missionRecap := c.getMissionRecap(rNewMission.GetMission())

			msg := helpers.NewMessage(c.ChatID, missionRecap)
			msg.ParseMode = tgbotapi.ModeHTML
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.accept")),
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.leave")),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
				),
			)
			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}

			// Recupero e salvo in payload il missionID
			c.Payload.MissionID = rNewMission.GetMission().GetID()
		}

		c.CurrentState.Stage = 1
	// Gestione di accetta o rifiuto missione
	case 1:
		// Missione accettata
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.accept") == c.Update.Message.Text {
			if _, err = config.App.Server.Connection.AcceptMission(helpers.NewContext(1), &pb.AcceptMissionRequest{
				PlayerID:  c.Player.GetID(),
				MissionID: c.Payload.MissionID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.accepted"))
			msg.ParseMode = tgbotapi.ModeHTML
			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}

			c.ForceBackTo = true
			// c.CurrentState.Completed = true
		} else {
			// Abbandono missione se avviata
			_, _ = config.App.Server.Connection.LeaveMission(helpers.NewContext(1), &pb.LeaveMissionRequest{
				PlayerID: c.Player.GetID(),
			})

			c.CurrentState.Stage = 0
			c.Stage()
		}

	// Il Player ha completato la missione
	case 2:
		// Effettuo chiamata al WS per recuperare reward del player
		var rGetMissionReward *pb.GetMissionRewardResponse
		if rGetMissionReward, err = config.App.Server.Connection.GetMissionReward(helpers.NewContext(1), &pb.GetMissionRewardRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		rewards := fmt.Sprintf("<b>+%v</b>💰\n<b>+%v</b>🏵",
			rGetMissionReward.GetMoney(),
			rGetMissionReward.GetExp(),
		)
		if rGetMissionReward.GetDiamond() > 0 {
			rewards += fmt.Sprintf("\n<b>+%v</b>💎", rGetMissionReward.GetDiamond())
		}

		msg := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.reward", rewards),
		)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}
}

// getMissionRecap
func (c *SafePlanetMissionSearchController) getMissionRecap(mission *pb.Mission) (missionRecap string) {
	var err error

	// In base alla categoria della missione costruisco il messaggio
	missionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.type",
		helpers.Trans(c.Player.Language.Slug,
			fmt.Sprintf("safeplanet.mission.type.%s", mission.GetMissionCategory().GetSlug()),
		),
		mission.ID,
	)

	// Recupero quotazione della missione allo stato attuale
	var rCheckMissionReward *pb.CheckMissionRewardResponse
	if rCheckMissionReward, err = config.App.Server.Connection.CheckMissionReward(helpers.NewContext(1), &pb.CheckMissionRewardRequest{
		MissionID: mission.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	rewards := fmt.Sprintf("+<b>%v</b>💰 +<b>%v</b>🏵",
		rCheckMissionReward.GetMoney(),
		rCheckMissionReward.GetExp(),
	)
	if rCheckMissionReward.GetDiamond() > 0 {
		rewards += fmt.Sprintf(" +<b>%v</b>💎", rCheckMissionReward.GetDiamond())
	}

	missionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.quotation", rewards)

	// Verifico quale tipologia di missione è stata estratta
	switch mission.GetMissionCategory().GetSlug() {
	// Trova il pianeta
	case "resources_finding":
		var missionPayload *pb.MissionResourcesFinding
		helpers.UnmarshalPayload(mission.GetPayload(), &missionPayload)

		// Recupero enitità risorsa da cercare
		var rGetResourceByID *pb.GetResourceByIDResponse
		if rGetResourceByID, err = config.App.Server.Connection.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
			ID: missionPayload.GetResourceID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		missionRecap += helpers.Trans(c.Player.Language.Slug,
			"safeplanet.mission.type.resources_finding.description",
			missionPayload.GetResourceQty(),
			helpers.GetResourceCategoryIcons(rGetResourceByID.GetResource().GetResourceCategoryID()),
			rGetResourceByID.GetResource().GetName(),
			rGetResourceByID.GetResource().GetRarity().GetSlug(),
			helpers.GetResourceBaseIcons(rGetResourceByID.GetResource().GetBase()),
		)

	// Trova le risorse
	case "planet_finding":
		var missionPayload *pb.MissionPlanetFinding
		helpers.UnmarshalPayload(mission.GetPayload(), &missionPayload)

		// Recupero pianeta da trovare
		var rGetPlanetByID *pb.GetPlanetByIDResponse
		if rGetPlanetByID, err = config.App.Server.Connection.GetPlanetByID(helpers.NewContext(1), &pb.GetPlanetByIDRequest{
			PlanetID: missionPayload.GetPlanetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		missionRecap += helpers.Trans(c.Player.Language.Slug,
			"safeplanet.mission.type.planet_finding.description",
			rGetPlanetByID.GetPlanet().GetName(),
			rGetPlanetByID.GetPlanet().GetPlanetSystem().GetName(),
			rGetPlanetByID.GetPlanet().GetPlanetSystem().GetID(),
		)

	// Uccidi il nemico
	case "kill_mob":
		var missionPayload *pb.MissionKillMob
		helpers.UnmarshalPayload(mission.GetPayload(), &missionPayload)

		// Recupero enemy da Uccidere
		var rGetEnemyByID *pb.GetEnemyByIDResponse
		if rGetEnemyByID, err = config.App.Server.Connection.GetEnemyByID(helpers.NewContext(1), &pb.GetEnemyByIDRequest{
			EnemyID: missionPayload.GetEnemyID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero pianeta di dove si trova il mob
		var rGetPlanetByMapID *pb.GetPlanetByMapIDResponse
		if rGetPlanetByMapID, err = config.App.Server.Connection.GetPlanetByMapID(helpers.NewContext(1), &pb.GetPlanetByMapIDRequest{
			MapID: rGetEnemyByID.GetEnemy().GetPlanetMapID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		missionRecap += helpers.Trans(c.Player.Language.Slug,
			"safeplanet.mission.type.kill_mob.description",
			rGetEnemyByID.GetEnemy().GetName(),
			rGetPlanetByMapID.GetPlanet().GetName(),
			rGetPlanetByMapID.GetPlanet().GetPlanetSystem().GetName(),
			rGetPlanetByMapID.GetPlanet().GetPlanetSystem().GetID(),
		)
	}

	return
}
