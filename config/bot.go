package config

import (
	"errors"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/joho/godotenv/autoload"
)

var (
	// TBot - Telegram bot
	TBot *tgbotapi.BotAPI
)

//BotUp - BotUp
func BotUp() {
	telegramAPIKey := os.Getenv("TELEGRAM_APIKEY")
	if telegramAPIKey == "" {
		ErrorHandler("$TELEGRAM_APIKEY must be set", errors.New("TelegramApiKey Missing"))
	}

	var err error
	TBot, err = tgbotapi.NewBotAPI(telegramAPIKey)
	if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
		TBot.Debug = true
	}

	if err != nil {
		ErrorHandler("tgbotapi.NewBotAPI(telegramAPIKey) return Error!", err)
	}

	log.Println("************************************************")
	log.Println("Bot connected: " + TBot.Self.UserName)
	log.Println("************************************************")
}
