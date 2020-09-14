package controllers

import (
	"fmt"
	"strconv"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetBankController
// ====================================
type SafePlanetBankController struct {
	Payload struct {
		Type string
	}
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetBankController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Player = player
	c.Update = update

	// Verifico se è impossibile inizializzare
	if !c.InitController(ControllerConfiguration{
		Controller: "route.safeplanet.bank",
		Payload:    c.Payload,
		ControllerBack: ControllerBack{
			To:        &MenuController{},
			FromStage: 0,
		},
	}) {
		return
	}

	// Carico payload
	if err = helpers.GetPayloadController(c.Player.ID, c.CurrentState.Controller, &c.Payload); err != nil {
		panic(err)
	}

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
	if err = c.Completing(&c.Payload); err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *SafePlanetBankController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false
	case 1:
		if helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit"),
			helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws"),
		}) {
			return false
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true
	case 2:
		// TODO: Verificare importo
		return false
	}

	return true
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetBankController) Stage() (err error) {
	switch c.CurrentState.Stage {
	// Invio messaggio con recap stats
	case 0:
		var infoBank string
		infoBank = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.info")

		msg := services.NewMessage(c.Player.ChatID, infoBank)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		msg.ParseMode = "HTML"
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Bank
		var rGetPlayerEconomy *pb.GetPlayerEconomyResponse
		if rGetPlayerEconomy, err = services.NnSDK.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
			PlayerID:    c.Player.GetID(),
			EconomyType: "bank",
		}); err != nil {
			return err
		}

		// Money
		var rGetPlayerEconomyMoney *pb.GetPlayerEconomyResponse
		if rGetPlayerEconomyMoney, err = services.NnSDK.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
			PlayerID:    c.Player.GetID(),
			EconomyType: "money",
		}); err != nil {
			return err
		}

		msg = services.NewMessage(c.Player.ChatID,
			helpers.Trans(
				c.Player.Language.Slug,
				"safeplanet.bank.account_details",
				rGetPlayerEconomyMoney.GetValue(),
				rGetPlayerEconomy.GetValue(),
			),
		)
		msg.ParseMode = "Markdown"
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1
	case 1:
		var mainMessage string
		var keyboardRowQuantities [][]tgbotapi.KeyboardButton
		c.Payload.Type = c.Update.Message.Text

		switch c.Payload.Type {
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit"):
			mainMessage = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit_message")

		case helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws"):
			mainMessage = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws_message")
		}

		// Inserisco le quantità di default per il prelievo/deposito
		for i := 1; i <= 5; i++ {
			keyboardRow := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(fmt.Sprintf("%d", i)),
			)
			keyboardRowQuantities = append(keyboardRowQuantities, keyboardRow)
		}

		// Aggiungo tasti back and clears
		keyboardRowQuantities = append(keyboardRowQuantities, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, mainMessage)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowQuantities,
		}
		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		// Se la validazione è passata vuol dire che è stato
		// inserito un importo valido e quindi posso eseguiore la transazione
		// in base alla tipologia scelta

		// Converto valore richiesto in int
		value, err := strconv.Atoi(c.Update.Message.Text)
		if err != nil {
			return err
		}

		var text string
		text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.operation_done")
		switch c.Payload.Type {
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit"):
			if _, err = services.NnSDK.BankDeposit(helpers.NewContext(1), &pb.BankDepositRequest{
				PlayerID: c.Player.ID,
				Amount:   int32(value),
			}); err != nil {
				text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.transaction_error")
			}
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws"):
			if _, err = services.NnSDK.BankWithdraw(helpers.NewContext(1), &pb.BankWithdrawRequest{
				PlayerID: c.Player.ID,
				Amount:   int32(value),
			}); err != nil {
				text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.transaction_error")
			}
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, text)
		msg.ParseMode = "markdown"

		if _, err = services.SendMessage(msg); err != nil {
			return err
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
