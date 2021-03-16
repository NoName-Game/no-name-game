package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetMissionController
// ====================================
type SafePlanetMissionController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMissionController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.mission",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 1,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
				2: {"route.breaker.clears", "route.breaker.continue"},
			},
		},
	}) {
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
	c.Completing(nil)
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetMissionController) Validator() (hasErrors bool) {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico avvio missione
	// ##################################################################################################
	case 0:
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.start") == c.Update.Message.Text {
			c.CurrentState.Stage = 1
		}

	// ##################################################################################################
	// In questo stage andremo a verificare lo stato della missione
	// ##################################################################################################
	case 2:
		// Verifico se il player ha deciso di abbandonare la missione
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.leave_mission") {
			if _, err = config.App.Server.Connection.LeaveMission(helpers.NewContext(1), &pb.LeaveMissionRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Stage = 0
			return
		}

		var rCheckMission *pb.CheckMissionResponse
		if rCheckMission, err = config.App.Server.Connection.CheckMission(helpers.NewContext(1), &pb.CheckMissionRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Verifico se realmente il player è in missione
		if !rCheckMission.GetInMission() {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.check_nomission")

			// Aggiungo anche abbandona
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
					),
				),
			)

			return true
		}

		// Verifico se il player ha completato la missione
		if !rCheckMission.GetCompleted() {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.check")

			// Aggiungo anche abbandona
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.continue"),
					),
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.leave_mission"),
					),
				),
			)

			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetMissionController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// Primo avvio chiedo al player se vuole avviare una nuova mission
	case 0:
		var keyboardRows [][]tgbotapi.KeyboardButton
		keyboardRows = append(keyboardRows, []tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.start")),
		})

		// Aggiungo anche abbandona
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.menu"),
			),
		))

		// Invio messaggi con il tipo di missioni come tastierino
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.info"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRows,
			ResizeKeyboard: true,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

	// In questo stage verrà recuperato il tempo di attesa per il
	// completamnto della missione e notificato al player
	case 1:
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

		// Invio messaggio di attesa
		msg := helpers.NewMessage(c.Player.ChatID, missionRecap)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stato
		c.CurrentState.Stage = 2
		c.ForceBackTo = true
	case 2:
		// Effettuo chiamata al WS per recuperare reward del player
		var rGetMissionReward *pb.GetMissionRewardResponse
		if rGetMissionReward, err = config.App.Server.Connection.GetMissionReward(helpers.NewContext(1), &pb.GetMissionRewardRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug,
				"safeplanet.mission.reward",
				rGetMissionReward.GetMoney(),
				rGetMissionReward.GetDiamond(),
				rGetMissionReward.GetExp(),
			),
		)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}

// getMissionRecap
func (c *SafePlanetMissionController) getMissionRecap(mission *pb.Mission) (missionRecap string) {
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

	missionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.quotation",
		rCheckMissionReward.GetMoney(),
		rCheckMissionReward.GetDiamond(),
		rCheckMissionReward.GetExp(),
	)

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
