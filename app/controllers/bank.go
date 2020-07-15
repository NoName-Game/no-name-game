package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// BankController
// ====================================
type BankController struct {
	Payload struct {
		Type string
	}
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *BankController) Handle(player nnsdk.Player, update tgbotapi.Update, proxy bool) {
	// Inizializzo variabili del controler
	var err error
	var playerStateProvider providers.PlayerStateProvider

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.safeplanet.bank",
		c.Payload,
		[]string{},
		player,
		update,
	) {
		return
	}

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &MenuController{}) {
			return
		}
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// Validate
	var hasError bool
	hasError, err = c.Validator()
	if err != nil {
		panic(err)
	}

	// Se ritornano degli errori
	if hasError {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
		// validatorMsg.ReplyMarkup = c.Validation.ReplyKeyboard

		_, err = services.SendMessage(validatorMsg)
		if err != nil {
			panic(err)
		}

		return
	}

	// Ok! Run!
	err = c.Stage()
	if err != nil {
		panic(err)
	}

	// Aggiorno stato finale
	payloadUpdated, _ := json.Marshal(c.Payload)
	c.State.Payload = string(payloadUpdated)
	c.State, err = playerStateProvider.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico completamento
	err = c.Completing()
	if err != nil {
		panic(err)
	}
}

// ====================================
// Validator
// ====================================
func (c *BankController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")
	c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
			),
		),
	)

	switch c.State.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err
	case 1:
		if helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit"),
			helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws"),
		}) {
			return false, err
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true, err
	case 2:
		// TODO: Verificare importo
		return false, err
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *BankController) Stage() (err error) {
	switch c.State.Stage {
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
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		var playerProvider providers.PlayerProvider
		money, _ := playerProvider.GetPlayerEconomy(c.Player.ID, "bank")

		msg = services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.account_details", money.Value))
		msg.ParseMode = "Markdown"
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}
		// Avanzo di stage
		c.State.Stage = 1
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
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.State.Stage = 2
	case 2:
		//var transactionProvider providers.TransactionProvider
		var npcProvider providers.NpcProvider
		// Se la validazione è passata vuol dire che è stato
		// inserito un importo valido e quindi posso eseguiore la transazione
		// in base alla tipologia scelta

		// Converto valore richiesto in int
		value, err := strconv.Atoi(c.Update.Message.Text)
		if err != nil {
			return err
		}
		var operationType uint
		switch c.Payload.Type {
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.deposit"):
			operationType = 0
		case helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.withdraws"):
			operationType = 1
		}
		request := nnsdk.BankActionRequest{
			OperationType: operationType,
			PlayerID:      c.Player.ID,
			Amount:        int32(value),
		}
		_, err = npcProvider.Bank(request)

		// Registro transazione
		/*_, err = transactionProvider.CreateTransaction(nnsdk.TransactionRequest{
			Value:                 int32(value),
			TransactionTypeID:     1,
			TransactionCategoryID: uint(transactionCategory),
			PlayerID:              c.Player.ID,
		})*/
		var text string
		if err != nil {
			// Errore nella transazione
			text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.transaction_error")
		} else {
			text = helpers.Trans(c.Player.Language.Slug, "safeplanet.bank.operation_done")
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, text)
		msg.ParseMode = "markdown"

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		*c.State.Completed = true
	}

	return
}
