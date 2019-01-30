package config

import (
	"log"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/joho/godotenv/autoload"
)

var (
	// TBot - Telegram bot
	TBot *tgbotapi.BotAPI
)

func init() {
	telegramAPIKey := os.Getenv("TELEGRAM_APIKEY")
	if telegramAPIKey == "" {
		log.Panicln("$TELEGRAM_APIKEY must be set")
	}

	var err error
	TBot, err = tgbotapi.NewBotAPI(telegramAPIKey)
	TBot.Debug = true
	if err != nil {
		log.Panic(err)
	}

	log.Println("************************************************")
	log.Println("Bot connected: " + TBot.Self.UserName)
	log.Println("************************************************")
}
