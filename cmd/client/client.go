package main

import (
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/router"
	"github.com/sirupsen/logrus"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Init
func init() {
	config.App.Bootstrap()
}

// Run - The Game!
func main() {
	var err error

	// Recupero stati/messaggio da telegram
	var updates tgbotapi.UpdatesChannel
	if updates, err = config.App.Bot.GetUpdates(); err != nil {
		logrus.Errorf("Error getting updates: %s", err.Error())
	}

	// Gestisco update ricevuti
	for update := range updates {
		go handleUpdate(update)
	}
}

// handleUpdate - Gestisco singolo update
func handleUpdate(update tgbotapi.Update) {
	var err error

	// Gestisco utente
	var player *pb.Player
	if player, err = helpers.HandleUser(update); err != nil {
		logrus.Panic(err)
	}

	// Gestico panic
	defer recoverUpdate(player, update)

	if _, err := helpers.GetAntiFlood(player.ID); err != nil {
		// Setto antiflood player
		_ = helpers.SetAntiFlood(player.ID)

		// Se non è stato trovato nulla allora posso eseguire il tutto
		router.Routing(player, update)

		// Cancello antiflood player
		_ = helpers.DelAntiFlood(player.ID)

		return
	}
}

func recoverUpdate(player *pb.Player, update tgbotapi.Update) {
	if err := recover(); err != nil {
		if err, ok := err.(error); ok {
			logrus.Errorf("[*] Recoverd Error: %v", err)
		}

		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := helpers.NewMessage(update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "validator.error"))
		validatorMsg.ParseMode = tgbotapi.ModeHTML
		validatorMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.clears")),
			),
		)

		_, _ = helpers.SendMessage(validatorMsg)
	}
}
