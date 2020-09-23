package main

import (
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/router"
	"github.com/sirupsen/logrus"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

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
		// Gestisco singolo update in worker dedicato
		go handleUpdate(update)
		// handleUpdate(update)
	}
}

// handleUpdate - Gestisco singolo update
func handleUpdate(update tgbotapi.Update) {
	var err error

	// Gestico panic
	defer func() {
		if err := recover(); err != nil {
			// TODO: Eseguire qualcosa se esplode male
			logrus.Errorf("[*] Recoverd Error: %v", err)
		}
	}()

	// Gestisco utente
	var player *pb.Player
	if player, err = helpers.HandleUser(update); err != nil {
		panic(err)
	}

	// Gestisco update
	router.Routing(player, update)
}
