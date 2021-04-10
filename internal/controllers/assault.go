package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Assault
// ====================================
type AssaultController struct {
	Controller
	Payload struct {
		DefenderID uint32
		DefenderInParty bool
	}
}

func (c *AssaultController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.assault",
			Payload: &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlanetController{},
				FromStage: 0,
			},
			PlanetType: []string{"default", "titan"},
			BreakerPerStage: map[int32][]string{
				1: {"route.breaker.menu"},
				2: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *AssaultController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	if c.Validator() {
		c.Validate()
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(&c.Payload)
}

func (c *AssaultController) Validator() bool {
	switch c.CurrentState.Stage {
	case 1:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "route.assault.scan.start") {
			return true
		}
	case 2:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.assault.scan.ingage") {
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "route.assault.scan.next") {
			c.CurrentState.Stage = 1
			return false
		}

		return true
	}

	return false
}

// FLOW: Scan -> Confirm -> Assault -> Reward
func (c *AssaultController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		// Chiedo se vuole scansionare
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "route.assault.info"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.assault.scan.start")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err.Error())
		}

		c.CurrentState.Stage = 1
	case 1:
		// Avvia la scansione e recupera player avversario
		var scanResult *pb.ScanPlanetResponse
		if scanResult, err = config.App.Server.Connection.Scan(helpers.NewContext(1), &pb.ScanPlanetRequest{PlayerID: c.Player.ID}); err != nil {
			if strings.Contains(err.Error(), "not enough fuel") {
				// Non ha più fuel, concludiamo
				msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "route.assault.error.no_fuel"))
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
					),
				)
				if _, err = helpers.SendMessage(msg); err != nil {
					c.Logger.Panic(err.Error())
				}
				c.CurrentState.Completed = true
				return
			} else if strings.Contains(err.Error(),"no players in planet") {
				// Non ci sono player
				msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "route.assault.error.no_player"))
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
					),
				)
				if _, err = helpers.SendMessage(msg); err != nil {
					c.Logger.Panic(err.Error())
				}
				c.CurrentState.Completed = true
				return
			}
			c.Logger.Panic(err.Error())
		}

		// Mi salvo il player da attaccare
		c.Payload.DefenderID = scanResult.GetPlayerID()
		c.Payload.DefenderInParty = scanResult.GetInParty()

		var rGetPlayerShip *pb.GetPlayerShipEquippedResponse
		if rGetPlayerShip, err = config.App.Server.Connection.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{PlayerID: scanResult.GetPlayerID()}); err != nil {
			c.Logger.Panic(err.Error())
		}

		// Costruisco il messaggio e chiedo all'utente se vuole effettuare l'attacco.
		var textCode string
		if scanResult.GetInParty() {
			textCode = "route.assault.scan.info_party"
		} else {
			textCode = "route.assault.scan.info_noparty"
		}
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, textCode, helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ship.category.%s", rGetPlayerShip.GetShip().GetShipCategory().GetSlug()))))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.assault.scan.ingage")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.assault.scan.next")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err.Error())
		}
		c.CurrentState.Stage = 2
	case 2:
		var attackerPartyID, defenderPartyID uint32
		// Il player ha deciso di attaccare il party avversario
		// Recupero party avversario e party player
		var rGetPlayerParty *pb.GetPartyDetailsResponse
		if rGetPlayerParty, err = config.App.Server.Connection.GetPartyDetails(helpers.NewContext(1), &pb.GetPartyDetailsRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			if strings.Contains(err.Error(),"player not in party") {
				attackerPartyID = 0
			} else {
				c.Logger.Panic(err.Error())
			}
		} else {
			attackerPartyID = rGetPlayerParty.GetPartyID()
		}
		defenderPartyID = 0
		if c.Payload.DefenderInParty {
			var rGetDefenderParty *pb.GetPartyDetailsResponse
			if rGetDefenderParty, err = config.App.Server.Connection.GetPartyDetails(helpers.NewContext(1), &pb.GetPartyDetailsRequest{
				PlayerID: c.Payload.DefenderID,
			}); err != nil {
				c.Logger.Panic(err.Error())
			} else {
				defenderPartyID = rGetDefenderParty.GetPartyID()
			}
		}

		var rStartAssault *pb.StartAssaultResponse
		if rStartAssault, err = config.App.Server.Connection.StartAssault(helpers.NewContext(1), &pb.StartAssaultRequest{
			AttackerID:      c.Player.ID,
			AttackerPartyID: attackerPartyID,
			DefenderID:      c.Payload.DefenderID,
			DefenderPartyID: defenderPartyID,
		}); err != nil {
			c.Logger.Panic(err.Error())
		}
		var winner string
		if rStartAssault.AttackerDefeated {
			winner = helpers.Trans(c.Player.Language.Slug, "route.assault.defender")
		} else {
			winner = helpers.Trans(c.Player.Language.Slug, "route.assault.attacker")
		}
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "route.assault.ingage.results", rStartAssault.GetTurns(), rStartAssault.GetAttackerTotalDamage(), rStartAssault.GetDefenderTotalDamage(), winner))
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err.Error())
		}

		c.CurrentState.Completed = true
	}
}
