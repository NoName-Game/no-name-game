package game

import (
	"gitlab.com/Valkyrie00/no-name/bot"
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

		msg := bot.NewMessage(update.Message.Chat.ID, update.Message.Text)
		bot.SendNewMessage(msg)
	}
}
