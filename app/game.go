package app

import (
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
		if update.Message != nil {
			if update.Message.From.UserName == "" {
				msg := bot.NewMessage(update.Message.Chat.ID, trans("miss_username", "en-US"))
				bot.SendMessage(msg)
				continue
			}

			routing(update)
		}
	}
}
