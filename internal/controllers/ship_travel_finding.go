package controllers

import (
	"fmt"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// ShipTravelFindingController
// ====================================
type ShipTravelFindingController struct {
	Controller
	Payload struct {
		StarNearestMapName  map[int]string
		CompleteWithDiamond bool
	}
}

// ====================================
// Handle
// ====================================
func (c *ShipTravelFindingController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.ship.travel.finding",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBlocked: []string{"exploration", "hunting"},
			ControllerBack: ControllerBack{
				To:        &ShipTravelController{},
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
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *ShipTravelFindingController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// In questo stage verifico che il player abbia pasasto la stella vicina
	// ##################################################################################################
	case 1:
		if !helpers.InArray(c.Update.Message.Text, c.Payload.StarNearestMapName) {
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

			return true
		}
	case 2:
		var err error
		var rCheckShipTravel *pb.CheckShipTravelResponse
		rCheckShipTravel, _ = config.App.Server.Connection.CheckShipTravel(helpers.NewContext(1), &pb.CheckShipTravelRequest{
			PlayerID: c.Player.ID,
		})

		// Il player sta già effettuando un viaggio
		if rCheckShipTravel.GetTravelInProgress() {
			c.CurrentState.Stage = 2

			// Se il viaggio non è finito
			if !rCheckShipTravel.GetFinishTraveling() {
				if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "complete_with_diamond") {
					c.Payload.CompleteWithDiamond = true
					c.CurrentState.Stage = 2
					return false
				}

				var finishAt time.Time
				if finishAt, err = helpers.GetEndTime(rCheckShipTravel.GetTravelingEndTime(), c.Player); err != nil {
					c.Logger.Panic(err)
				}

				// Calcolo diamanti del player
				var rGetPlayerEconomyDiamond *pb.GetPlayerEconomyResponse
				if rGetPlayerEconomyDiamond, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
					PlayerID:    c.Player.GetID(),
					EconomyType: pb.GetPlayerEconomyRequest_DIAMOND,
				}); err != nil {
					c.Logger.Panic(err)
				}

				// Invio messaggio recap fine viaggio
				c.Validation.Message = helpers.Trans(
					c.Player.Language.Slug,
					"ship.travel.wait",
					finishAt.Format("15:04:05"),
					rGetPlayerEconomyDiamond.GetValue(),
				)

				// Aggiungi possibilità di velocizzare
				c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							helpers.Trans(c.Player.Language.Slug, "complete_with_diamond"),
						),
					),
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							helpers.Trans(c.Player.Language.Slug, "route.breaker.continue"),
						),
					),
				)

				return true
			}
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *ShipTravelFindingController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// In questo stage recupero le stelle più vicine disponibili per il player
	case 0:
		// Recupero nave attualemente attiva
		var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
		if rGetPlayerShipEquipped, err = config.App.Server.Connection.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero informazioni di viaggi
		var rShipTravelInfo *pb.ShipTravelInfoResponse
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.ship.travel.favorite") {
			if rShipTravelInfo, err = config.App.Server.Connection.ShipTravelFavoriteInfo(helpers.NewContext(1), &pb.ShipTravelInfoRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}
		} else {
			if rShipTravelInfo, err = config.App.Server.Connection.ShipTravelRadarInfo(helpers.NewContext(1), &pb.ShipTravelInfoRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}
		}

		var starNearestMapName = make(map[int]string)
		var starNearestMapInfo = make(map[int]*pb.ShipTravelInfoResponse_ShipTravelInfo)
		var msgNearestStars string

		// Keyboard con riassunto risorse necessarie
		var keyboardRowStars [][]tgbotapi.KeyboardButton
		for _, explorationInfo := range rShipTravelInfo.GetInfo() {
			// Per scontato che il pianeta sia raggiungibile
			reachable := true

			// Verifico se il pianeta è raggiungibile in base allo stato della nave
			reachableMsg := helpers.Trans(c.Player.Language.Slug, "ship.travel.reachable")
			if rGetPlayerShipEquipped.GetShip().GetIntegrity() < explorationInfo.Integrity ||
				rGetPlayerShipEquipped.GetShip().GetTank() < explorationInfo.Fuel {
				reachable = false
				reachableMsg = helpers.Trans(c.Player.Language.Slug, "ship.travel.unreachable")
			}

			// Se il pianeta è sicuro allora appendo al nome l'icona di riferimento
			planetName := explorationInfo.Planet.Name
			if explorationInfo.Planet.Safe {
				planetName = fmt.Sprintf("%s 🏟", explorationInfo.Planet.Name)
			}

			// Calcolo tempo di esplorazione e se il viaggi è più breve di 1 ora riporto in minuti
			travelTime := fmt.Sprintf("%v (%s)", explorationInfo.Time/60, helpers.Trans(c.Player.Language.Slug, "hours"))
			if explorationInfo.Time/60 <= 0 {
				travelTime = fmt.Sprintf("%v (%s)", explorationInfo.Time, helpers.Trans(c.Player.Language.Slug, "minutes"))
			}

			// Mostro se il pianeta è stato mai raggiunto
			var planetAlreadyVisited string
			if explorationInfo.AlreadyVisited {
				planetAlreadyVisited = "❇️"
			}

			msgNearestStars += fmt.Sprintf("\n\n🌏 %s - 💫 %s (%d) %s\n%s ⏱ %v ⛽️ -%v%% 🔧 -%v%%",
				planetName, explorationInfo.Planet.PlanetSystem.Name, explorationInfo.Planet.PlanetSystem.ID, planetAlreadyVisited,
				reachableMsg,
				travelTime,
				explorationInfo.Fuel,
				explorationInfo.Integrity,
			)

			// Aggiungo per la validazione
			starNearestMapName[int(explorationInfo.Planet.ID)] = explorationInfo.Planet.Name
			starNearestMapInfo[int(explorationInfo.Planet.ID)] = explorationInfo

			// Aggiungo stelle raggiungibili alla keyboard
			if reachable {
				keyboardRowStars = append(keyboardRowStars, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
					explorationInfo.Planet.Name,
				)))
			}
		}

		keyboardRowStars = append(keyboardRowStars,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s %s",
				helpers.Trans(c.Player.Language.Slug, "ship.travel.research", len(rShipTravelInfo.GetInfo())),
				msgNearestStars,
			),
		)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowStars,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Update state
		c.Payload.StarNearestMapName = starNearestMapName
		c.CurrentState.Stage = 1

	// Verifico quale stella ha scelto il player e mando messaggio indicando il tempo
	// necessario al suo raggiungimento
	case 1:
		var rStartShipTravel *pb.StartShipTravelResponse
		if rStartShipTravel, err = config.App.Server.Connection.StartShipTravel(helpers.NewContext(1), &pb.StartShipTravelRequest{
			PlayerID:   c.Player.GetID(),
			PlanetName: c.Update.Message.Text,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero orario fine viaggio
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rStartShipTravel.GetTravelingEndTime(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "ship.travel.exploring", finishAt.Format("15:04:05 01/02")),
		)
		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
		c.ForceBackTo = true

	// Fine esplorazione
	case 2:
		// Verifico se ha gemmato
		if c.Payload.CompleteWithDiamond {
			if _, err := config.App.Server.Connection.EndShipTravelDiamond(helpers.NewContext(1), &pb.EndShipTravelRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				// Messaggio errore completamento
				msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "ship.travel.complete_diamond_error"))
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
					),
				)

				if _, err = helpers.SendMessage(msg); err != nil {
					c.Logger.Panic(err)
				}

				// Fondamentale, esco senza chiudere
				c.ForceBackTo = true
				return
			}
		} else {
			// Costruisco chiamata per aggiornare posizione e scalare il quantitativo di carburante usato
			if _, err := config.App.Server.Connection.EndShipTravel(helpers.NewContext(1), &pb.EndShipTravelRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "ship.travel.end"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Forzo cancellazione posizione player in cache
		_ = helpers.DelPlayerPlanetPositionInCache(c.Player.GetID())

		// Completo lo stato
		c.CurrentState.Completed = true
		// Imposto nil in modo da esser portato al menù principale
		c.Configurations.ControllerBack.To = nil
	}

	return
}
