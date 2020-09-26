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

	// Gestico panic
	defer recoverUpdate()

	// Gestisco utente
	var player *pb.Player
	if player, err = helpers.HandleUser(update); err != nil {
		logrus.Panic(err)
	}

	// Gestisco update
	router.Routing(player, update)
}

func recoverUpdate() {
	if err := recover(); err != nil {
		// TODO: Eseguire qualcosa se esplode male
		if err, ok := err.(error); ok {
			logrus.Errorf("[*] Recoverd Error: %v", err)
		}
	}
}
