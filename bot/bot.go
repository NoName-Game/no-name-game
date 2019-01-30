package bot

import (
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"gitlab.com/Valkyrie00/no-name/config"
)

// GetUpdates - return new updates
func GetUpdates() tgbotapi.UpdatesChannel {
	u := NewUpdate(0)
	u.Timeout = 60

	return config.TBot.GetUpdatesChan(u)
}

// SendNewMessage - send new message
func SendNewMessage(chattable tgbotapi.MessageConfig) tgbotapi.Message {
	message, err := config.TBot.Send(chattable)
	if err != nil {
		log.Println("Cant send message.")
	}

	return message
}
