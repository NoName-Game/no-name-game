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
			Controller: "route.safeplanet.mission",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 1,
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
		var rCheckMission *pb.CheckMissionResponse
		if rCheckMission, err = config.App.Server.Connection.CheckMission(helpers.NewContext(1), &pb.CheckMissionRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Verifico se realmente il player è in missione
		if !rCheckMission.GetInMission() {
			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"safeplanet.mission.check_nomission",
			)

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
			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"safeplanet.mission.check",
			)

			// Aggiungo anche abbandona
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.continue"),
					),
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.clears"),
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
				helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
			),
		))

		// Invio messaggi con il tipo di missioni come tastierino
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.info"))
		msg.ParseMode = "markdown"
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
		var rGetMission *pb.GetMissionResponse
		if rGetMission, err = config.App.Server.Connection.GetMission(helpers.NewContext(1), &pb.GetMissionRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// In base alla categoria della missione costruisco il messaggio
		var missionRecap string
		missionRecap += helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.type",
			helpers.Trans(c.Player.Language.Slug,
				fmt.Sprintf("safeplanet.mission.type.%s", rGetMission.GetMission().GetMissionCategory().GetSlug()),
			),
		)

		switch rGetMission.GetMission().GetMissionCategory().GetSlug() {
		case "resources_finding":
			var missionPayload *pb.MissionResourcesFinding
			helpers.UnmarshalPayload(rGetMission.GetMission().GetPayload(), &missionPayload)

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
				rGetResourceByID.GetResource().GetName(),
			)
		case "planet_finding":
			var missionPayload *pb.MissionPlanetFinding
			helpers.UnmarshalPayload(rGetMission.GetMission().GetPayload(), &missionPayload)

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
			)
		case "kill_mob":
			var missionPayload *pb.MissionKillMob
			helpers.UnmarshalPayload(rGetMission.GetMission().GetPayload(), &missionPayload)

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
				MapID: rGetEnemyByID.GetEnemy().GetMapID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			missionRecap += helpers.Trans(c.Player.Language.Slug,
				"safeplanet.mission.type.kill_mob.description",
				rGetEnemyByID.GetEnemy().GetName(),
				rGetPlanetByMapID.GetPlanet().GetName(),
			)
		}

		// Invio messaggio di attesa
		msg := helpers.NewMessage(c.Player.ChatID, missionRecap)
		msg.ParseMode = "markdown"
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
		msg.ParseMode = "markdown"
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
