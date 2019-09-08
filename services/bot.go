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
func GetUpdates() (tgbotapi.UpdatesChannel, error) {
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

// EditMessage - edit message
func NewEditMessage(chatID int64, messageID int, text string) tgbotapi.EditMessageTextConfig {
	return tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    chatID,
			MessageID: messageID,
		},
		Text: text,
	}
}

// SendMessage - send message
func SendMessage(chattable tgbotapi.Chattable) tgbotapi.Message {
	message, err := botAPI.Send(chattable)
	if err != nil {
		log.Println("Cant send message.")
	}

	return message
}

func NewAnswer(callbackQueryID string, text string, alert bool) tgbotapi.CallbackConfig {
	return tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQueryID,
		Text:            text,
		ShowAlert:       alert,
	}
}

func AnswerCallbackQuery(config tgbotapi.CallbackConfig) {
	_, err := botAPI.AnswerCallbackQuery(config)
	if err != nil {
		log.Println("Cant send answer.")
	}
}
