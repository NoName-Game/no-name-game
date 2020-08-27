package controllers

import (
	"fmt"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetMissionController
// ====================================
type SafePlanetMissionController struct {
	BaseController
	Payload struct {
		MissionID uint32
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMissionController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller: "route.safeplanet.mission",
		ControllerBack: ControllerBack{
			To:        &SafePlanetCoalitionController{},
			FromStage: 1,
		},
		Payload: c.Payload,
	}) {
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.PlayerData.CurrentState.Payload, &c.Payload)

	// Validate
	var hasError bool
	if hasError = c.Validator(); hasError {
		c.Validate()
		return
	}

	// Ok! Run!
	if err = c.Stage(); err != nil {
		panic(err)
	}

	// Completo progressione
	if err = c.Completing(c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetMissionController) Validator() (hasErrors bool) {
	switch c.PlayerData.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false

	// Verifico se ha scelto di avviare una nuova missione
	case 1:
		if helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.start") == c.Update.Message.Text {
			return false
		}

		return true

	// In questo stage andremo a verificare lo stato della missione
	case 2:
		var rCheckMission *pb.CheckMissionResponse
		rCheckMission, _ = services.NnSDK.CheckMission(helpers.NewContext(1), &pb.CheckMissionRequest{
			PlayerID:  c.Player.GetID(),
			MissionID: c.Payload.MissionID,
		})

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

		return false

	default:
		// Stato non riconosciuto ritorno errore
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.state")
	}

	// Ritorno errore generico
	return true
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetMissionController) Stage() (err error) {
	switch c.PlayerData.CurrentState.Stage {
	// Primo avvio chiedo al player se vuole avviare una nuova mission
	case 0:
		var keyboardRows [][]tgbotapi.KeyboardButton
		keyboardRows = append(keyboardRows, []tgbotapi.KeyboardButton{
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.start")),
		})

		// Aggiungo anche abbandona
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		))

		// Invio messaggi con il tipo di missioni come tastierino
		msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.mission.info"))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRows,
			ResizeKeyboard: true,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return
		}

		// Avanzo di stage
		c.PlayerData.CurrentState.Stage = 1

	// In questo stage verrà recuperato il tempo di attesa per il
	// completamnto della missione e notificato al player
	case 1:
		// Chiamo il ws e recupero il tipo di missione da effettuare
		// attraverso il tipo di missione costruisco il corpo del messaggio
		var rGetMission *pb.GetMissionResponse
		rGetMission, err = services.NnSDK.GetMission(helpers.NewContext(1), &pb.GetMissionRequest{
			PlayerID: c.Player.GetID(),
		})
		if err != nil {
			return err
		}

		// In base alla categoria della missione costruisco il messaggio
		var missionRecap string
		missionRecap += helpers.Trans(c.Player.Language.Slug,
			"safeplanet.mission.type",
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
			rGetResourceByID, err = services.NnSDK.GetResourceByID(helpers.NewContext(1), &pb.GetResourceByIDRequest{
				ID: missionPayload.GetResourceID(),
			})
			if err != nil {
				panic(err)
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
			rGetPlanetByID, err = services.NnSDK.GetPlanetByID(helpers.NewContext(1), &pb.GetPlanetByIDRequest{
				PlanetID: missionPayload.GetPlanetID(),
			})
			if err != nil {
				panic(err)
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
			rGetEnemyByID, err = services.NnSDK.GetEnemyByID(helpers.NewContext(1), &pb.GetEnemyByIDRequest{
				ID: missionPayload.GetEnemyID(),
			})
			if err != nil {
				panic(err)
			}

			// Recupero pianeta di dove si trova il mob
			var rGetPlanetByMapID *pb.GetPlanetByMapIDResponse
			rGetPlanetByMapID, err = services.NnSDK.GetPlanetByMapID(helpers.NewContext(1), &pb.GetPlanetByMapIDRequest{
				MapID: rGetEnemyByID.GetEnemy().GetMapID(),
			})
			if err != nil {
				panic(err)
			}

			missionRecap += helpers.Trans(c.Player.Language.Slug,
				"safeplanet.mission.type.kill_mob.description",
				rGetEnemyByID.GetEnemy().GetName(),
				rGetPlanetByMapID.GetPlanet().GetName(),
			)
		}

		// Invio messaggio di attesa
		msg := services.NewMessage(c.Player.ChatID,
			missionRecap,
		)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return
		}

		// Avanzo di stato
		c.Payload.MissionID = rGetMission.GetMission().GetID()
		c.PlayerData.CurrentState.Stage = 2
		c.ForceBackTo = true

	case 2:
		// Effettuo chiamata al WS per recuperare reward del player
		var rGetMissionReward *pb.GetMissionRewardResponse
		rGetMissionReward, err = services.NnSDK.GetMissionReward(helpers.NewContext(1), &pb.GetMissionRewardRequest{
			PlayerID:  c.Player.GetID(),
			MissionID: c.Payload.MissionID,
		})
		if err != nil {
			panic(err)
		}

		msg := services.NewMessage(c.Player.ChatID,
			helpers.Trans(c.Player.Language.Slug,
				"safeplanet.mission.reward",
				rGetMissionReward.GetMoney(),
				rGetMissionReward.GetDiamond(),
				rGetMissionReward.GetExp(),
			),
		)
		msg.ParseMode = "markdown"

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "exploration.continue")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "exploration.comeback")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.PlayerData.CurrentState.Completed = true
	}

	return
}
