package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetBankController
// ====================================
type SafePlanetBankController struct {
	Payload struct {
		Type string
	}
	Controller
}

func (c *SafePlanetBankController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.bank",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &MenuController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
				2: {"route.breaker.menu", "route.breaker.back"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetBankController) Handle(player *pb.Player, update tgbotapi.Update) {
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

// ====================================
// Validator
// ====================================
func (c *SafePlanetBankController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico tipologia di transazione
	// ##################################################################################################
	case 0:
		if helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit"),
			helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws"),
		}) {
			c.CurrentState.Stage = 1
		}
	case 2:
		// TODO: Verificare importo
		return false
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetBankController) Stage() {

	var err error
	switch c.CurrentState.Stage {
	// Invio messaggio con recap stats
	case 0:
		// Bank
		var rGetPlayerEconomy *pb.GetPlayerEconomyResponse
		if rGetPlayerEconomy, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
			PlayerID:    c.Player.GetID(),
			EconomyType: pb.GetPlayerEconomyRequest_BANK,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Money
		var rGetPlayerEconomyMoney *pb.GetPlayerEconomyResponse
		if rGetPlayerEconomyMoney, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
			PlayerID:    c.Player.GetID(),
			EconomyType: pb.GetPlayerEconomyRequest_MONEY,
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s\n\n%s",
			helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.info"),
			helpers.Trans(
				c.Player.Language.Slug,
				"safeplanet.bank.account_details",
				rGetPlayerEconomyMoney.GetValue(),
				rGetPlayerEconomy.GetValue(),
			),
		))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

	case 1:
		var mainMessage string
		var keyboardRowQuantities [][]tgbotapi.KeyboardButton
		switch c.Update.Message.Text {
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit"):
			mainMessage = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit_message")

			c.Payload.Type = "deposit"
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws"):
			mainMessage = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws_message")

			c.Payload.Type = "withdraws"
		}

		// Inserisco le quantità di default per il prelievo/deposito
		for i := 1; i <= 5; i++ {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(fmt.Sprintf("%d%%", i*5)),
			)
			keyboardRowQuantities = append(keyboardRowQuantities, keyboardRow)
		}

		// Aggiungo tasti back and clears
		keyboardRowQuantities = append(keyboardRowQuantities, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID, mainMessage)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowQuantities,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		// Se la validazione è passata vuol dire che è stato
		// inserito un importo valido e quindi posso eseguiore la transazione
		// in base alla tipologia scelta
		// Controllo in primis che non ci sia il % che indica un tipo diverso (accettiamo sia percentuali sia valori diretti)
		var value int
		if strings.Contains(c.Update.Message.Text, "%") {
			var percentage int
			if percentage, err = strconv.Atoi(strings.ReplaceAll(c.Update.Message.Text, "%", "")); err != nil {
				c.Logger.Panic(err)
			}
			switch c.Payload.Type {
			case "deposit":
				// Money
				var rGetPlayerEconomyMoney *pb.GetPlayerEconomyResponse
				if rGetPlayerEconomyMoney, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
					PlayerID:    c.Player.GetID(),
					EconomyType: pb.GetPlayerEconomyRequest_MONEY,
				}); err != nil {
					c.Logger.Panic(err)
				}
				value = (int(rGetPlayerEconomyMoney.GetValue()) * percentage) / 100
			case "withdraws":
				// Bank
				var rGetPlayerEconomy *pb.GetPlayerEconomyResponse
				if rGetPlayerEconomy, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
					PlayerID:    c.Player.GetID(),
					EconomyType: pb.GetPlayerEconomyRequest_BANK,
				}); err != nil {
					c.Logger.Panic(err)
				}
				value = (int(rGetPlayerEconomy.GetValue()) * percentage) / 100
			}
		} else {
			// Converto valore richiesto in int
			if value, err = strconv.Atoi(c.Update.Message.Text); err != nil {
				c.Logger.Panic(err)
			}
		}

		var text string
		text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.operation_done")

		switch c.Payload.Type {
		case "deposit":
			if _, err = config.App.Server.Connection.BankDeposit(helpers.NewContext(1), &pb.BankDepositRequest{
				PlayerID: c.Player.ID,
				Amount:   int32(value),
			}); err != nil {
				text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.transaction_error")
			}
		case "withdraws":
			if _, err = config.App.Server.Connection.BankWithdraw(helpers.NewContext(1), &pb.BankWithdrawRequest{
				PlayerID: c.Player.ID,
				Amount:   int32(value),
			}); err != nil {
				text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.transaction_error")
			}
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID, text)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
		c.Configurations.ControllerBack.To = &SafePlanetBankController{}
	}

	return
}
