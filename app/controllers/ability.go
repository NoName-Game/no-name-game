package controllers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// AbilityController
// ====================================
type AbilityController struct {
	Payload struct{}
	BaseController
}

var (
	AbilityLists = []string{
		"Strength",
		// "dexterity",
		// "constitution",
		"Intelligence",
		// "wisdom",
		// "charisma",
	}
)

// ====================================
// Handle
// ====================================
func (c *AbilityController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	var playerStateProvider providers.PlayerStateProvider

	c.Controller = "route.ability"
	c.Player = player
	c.Update = update

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	if c.Clear() {
		return
	}

	// Verifico se vuole tornare indietro di stato
	if c.BackTo(0) {
		new(MenuController).Handle(c.Player, c.Update)
		return
	}

	if c.Clear() {
		return
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
func (c *AbilityController) Validator() (hasErrors bool, err error) {
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
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "ability.back") {
			c.State.Stage = 0
			return false, err
		}

		// Verifico se l'abilità passata esiste nelle abilità censite e se il player ha punti disponibili
		for _, ability := range AbilityLists {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ability.%s", strings.ToLower(ability))) == c.Update.Message.Text {
				if *c.Player.Stats.AbilityPoint == 0 {
					c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "ability.no_point_left")
					return true, err
				}

				return false, err
			}
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")
		return true, err
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *AbilityController) Stage() (err error) {
	var playerStatsProvider providers.PlayerStatsProvider

	switch c.State.Stage {
	// Invio messaggio con recap stats
	case 0:
		var recapStats string
		recapStats = helpers.Trans(c.Player.Language.Slug, "ability.stats.type")

		// Recupero dinamicamente i valory delle statistiche per poi ciclarli con quelli consentiti
		rv := reflect.ValueOf(&c.Player.Stats)
		rv = rv.Elem()

		for _, ability := range AbilityLists {
			playerStat := rv.FieldByName(ability)
			fieldName := helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ability.%s", strings.ToLower(ability)))
			recapStats += fmt.Sprintf("<code>%-15v:%v</code>\n", fieldName, playerStat)
		}

		// Mostro quanti punti ha a disposizione il player
		messagePlayerTotalPoint := helpers.Trans(c.Player.Language.Slug, "ability.stats.total_point", *c.Player.Stats.AbilityPoint)

		// Creo tastierino con i soli componienti abilitati dal client
		var keyboardRow [][]tgbotapi.KeyboardButton
		for _, ability := range AbilityLists {
			row := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ability.%s", strings.ToLower(ability))),
				),
			)
			keyboardRow = append(keyboardRow, row)
		}

		// Aggiungo bottone cancella
		keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
			),
		))

		msg := services.NewMessage(c.Player.ChatID, fmt.Sprintf("%s\n\n%s", messagePlayerTotalPoint, recapStats))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			Keyboard:       keyboardRow,
			ResizeKeyboard: true,
		}
		msg.ParseMode = "HTML"
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Avanzo di stage
		c.State.Stage = 1
	case 1:
		// Incremento statistiche e aggiorno
		for _, ability := range AbilityLists {
			abilityName := helpers.Trans(c.Player.Language.Slug, "ability."+strings.ToLower(ability))

			if abilityName == c.Update.Message.Text {
				f := reflect.ValueOf(&c.Player.Stats).Elem().FieldByName(ability)
				f.SetUint(uint64(f.Interface().(uint) + 1))

				*c.Player.Stats.AbilityPoint--
			}
		}

		// Aggiorno statistiche player
		_, err = playerStatsProvider.UpdatePlayerStats(c.Player.Stats)
		if err != nil {
			return err
		}

		// Invio Messaggio di incremento abilità
		text := helpers.Trans(c.Player.Language.Slug, "ability.stats.completed", c.Update.Message.Text)
		msg := services.NewMessage(c.Player.ChatID, text)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}
	}

	return
}
