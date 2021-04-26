package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	"fmt"
	"strings"
	"time"

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
				To:        &ShipController{},
				FromStage: 0,
			},
			PlanetType: []string{"default", "titan"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
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
		return
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(&c.Payload)
}

func (c *AssaultController) Validator() bool {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		// Controllo che la nave sia integra
		// Recupero nave attualemente attiva
		var rGetPlayerShipEquipped *pb.GetPlayerShipEquippedResponse
		if rGetPlayerShipEquipped, err = config.App.Server.Connection.GetPlayerShipEquipped(helpers.NewContext(1), &pb.GetPlayerShipEquippedRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		if rGetPlayerShipEquipped.GetShip().GetIntegrity() == 0 {
			// Non può startare l'attività
			c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "route.assault.error.no_integrity")
			return true
		}
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
		// Fase di stampa.
		// Turno X:
		//		Il tuo PARTY ha inflitto: X danni e subito: X danni.
		// Recap Party:
		//		🚀 NAME 💨: ▰▰▰▰▰▱▱▱▱▱ 50%
		//		🚀 NAME 💨: ▰▰▰▰▰▱▱▱▱▱ 50%
		//		🚀 NAME 💨: ▰▰▰▰▰▱▱▱▱▱ 50%
		var rGetFormation *pb.GetFormationResponse
		if rGetFormation, err = config.App.Server.Connection.GetFormation(helpers.NewContext(1), &pb.GetFormationRequest{
			PlayerID: c.Player.ID,
			PartyID:  attackerPartyID,
		}); err != nil {
			c.Logger.Panic(err.Error())
		}
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "route.assault.ingage"))
		var msg_sended tgbotapi.Message
		if msg_sended, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err.Error())
		}
		total_recap := ""
		for i := uint32(1); i < rStartAssault.GetTurns()+1; i++ {
			// Recap turno
			damagePerTurnAttacker := rStartAssault.GetAttackerTotalDamage()/float64(rStartAssault.GetTurns())
			damagePerTurnDefender := rStartAssault.GetDefenderTotalDamage()/float64(rStartAssault.GetTurns())

			recap := "%s\n%s"
			turn_recap := helpers.Trans(c.Player.Language.Slug, "ruote.assault.turn_recap", i, damagePerTurnAttacker, damagePerTurnDefender)
			total_recap += turn_recap+"\n"
			party_recap := "<b>Party</b>:\n"
			for _, ship := range rGetFormation.Formation {
				party_recap += helpers.Trans(c.Player.Language.Slug, "route.assault.ship_status", ship.GetName()[0:4]+"...", helpers.GetShipCategoryIcons(ship.GetShipCategoryID()), helpers.GenerateHealthBar(ship.GetIntegrity()), ship.GetIntegrity())
				party_recap += "\n"
			}

			edit := helpers.NewEditMessage(c.Player.ChatID, msg_sended.MessageID, fmt.Sprintf(recap, turn_recap, party_recap))
			edit.ParseMode = tgbotapi.ModeHTML
			if _, err = helpers.SendMessage(edit); err != nil {
				c.Logger.Panic(err.Error())
			}
			time.Sleep(3 * time.Second)
		}
		var text string
		if rStartAssault.AttackerDefeated {
			text = helpers.Trans(c.Player.Language.Slug, "route.assault.end_defeat", total_recap)
		} else {
			text = helpers.Trans(c.Player.Language.Slug, "route.assault.end_win", total_recap)
		}
		msg = helpers.NewMessage(c.Player.ChatID, text)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err.Error())
		}

		c.CurrentState.Completed = true
	}
}
