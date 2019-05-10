package services

import (
	"errors"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/joho/godotenv/autoload" // Autload .env
)

var (
	// BotAPI - Telegram bot
	botAPI *tgbotapi.BotAPI
)

//BotUp - BotUp
func BotUp() {
	telegramAPIKey := os.Getenv("TELEGRAM_APIKEY")
	if telegramAPIKey == "" {
		ErrorHandler("$TELEGRAM_APIKEY must be set", errors.New("TelegramApiKey Missing"))
	}

	var err error
	botAPI, err = tgbotapi.NewBotAPI(telegramAPIKey)
	if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
		botAPI.Debug = true
	}

	if err != nil {
		ErrorHandler("tgbotapi.NewBotAPI(telegramAPIKey) return Error!", err)
	}

	log.Println("************************************************")
	log.Println("Bot connected: " + botAPI.Self.UserName)
	log.Println("************************************************")
}

// GetUpdates - return new updates
func GetUpdates() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return botAPI.GetUpdatesChan(u)
}

// NewMessage creates a new Message.
//
// chatID is where to send it, text is the message text.
func NewMessage(chatID int64, text string) tgbotapi.MessageConfig {
	return tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           chatID,
			ReplyToMessageID: 0,
		},
		Text:                  text,
		DisableWebPagePreview: false,
	}
}

// SendMessage - send message
func SendMessage(chattable tgbotapi.MessageConfig) tgbotapi.Message {
	message, err := botAPI.Send(chattable)
	if err != nil {
		log.Println("Cant send message.")
	}

	return message
}
