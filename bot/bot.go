package bot

import (
	"log"

	"bitbucket.org/no-name-game/no-name/config"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// GetUpdates - return new updates
func GetUpdates() tgbotapi.UpdatesChannel {
	u := NewUpdate(0)
	u.Timeout = 60

	return config.TBot.GetUpdatesChan(u)
}

// SendMessage - send message
func SendMessage(chattable tgbotapi.MessageConfig) tgbotapi.Message {
	message, err := config.TBot.Send(chattable)
	if err != nil {
		log.Println("Cant send message.")
	}

	return message
}
