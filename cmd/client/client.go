package main

import (
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/init/bootstrap"
	"bitbucket.org/no-name-game/nn-telegram/internal/router"

	"bitbucket.org/no-name-game/nn-telegram/init/logging"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Init
func init() {
	// Inizializzo servizi bot
	var err = bootstrap.Bootstrap()
	if err != nil {
		// Nel caso in cui uno dei servizi principale
		// dovesse entrare in errore in questo caso è meglio panicare
		panic(err)
	}
}

// Run - The Game!
func main() {
	var err error

	// Recupero stati/messaggio da telegram
	updates, err := config.App.Bot.GetUpdates()
	if err != nil {
		logging.ErrorHandler("Update channel error", err)
	}

	// Gestisco update ricevuti
	for update := range updates {
		// Gestisco singolo update in worker dedicato
		// go handleUpdate(update)
		handleUpdate(update)
	}
}

// handleUpdate - Gestisco singolo update
func handleUpdate(update tgbotapi.Update) {
	// Differisco controllo panic/recover
	// defer func() {
	// 	// Nel caso in cui panicasse
	// 	if err := recover(); err != nil {
	// 		// Registro errore
	// 		services.ErrorHandler("recover handle update", err.(error))
	// 	}
	// }()

	var err error
	// Gestisco utente
	var player *pb.Player
	player, err = helpers.HandleUser(update)
	if err != nil {
		panic(err)
	}

	// Gestisco update
	router.Routing(player, update)
}
