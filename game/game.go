package game

import (
	"log"

	"bitbucket.org/no-name-game/no-name/bot"
)

var (
	player Player
)

func init() {
	bootstrap()
}

// Run - The Game!
func Run() {
	updates := bot.GetUpdates()
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if update.Message.From.UserName == "" {
			msg := bot.NewMessage(update.Message.Chat.ID, "Non puoi continuare senza useraname")
			bot.SendMessage(msg)
			continue
		}

		// Block to test DB
		// First or create
		player.findByUsername(update.Message.From.UserName)
		if player.ID < 1 {
			player = Player{Username: update.Message.From.UserName}
			player.create()
		}

		log.Println(player)
		// ***************************

		routing(update)

		// msg := bot.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// bot.SendMessage(msg)
	}
}
