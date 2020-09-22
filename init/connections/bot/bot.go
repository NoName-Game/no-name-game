package bot

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/joho/godotenv/autoload" // Autload .env
)

// var (
// 	// BotAPI - Telegram bot
// 	botAPI *tgbotapi.BotAPI
// )

// Bot
type Bot struct {
	API *tgbotapi.BotAPI
}

// BotUp - Metodo per la connessione ai server telegram
func (bot *Bot) Init() {
	var err error

	// Recupero da env chiave telegram
	telegramAPIKey := os.Getenv("TELEGRAM_APIKEY")
	if telegramAPIKey == "" {
		panic("telegram ApiKey missing")
	}

	// Istanzio comunicazione con il servizio dedicato
	bot.API, err = tgbotapi.NewBotAPI(telegramAPIKey)
	if err != nil {
		panic(err)
	}

	// Nel caso in cui fosse in ambiente di sviluppo abilito il debug
	if appDebug := os.Getenv("TELEGRAM_DEBUG"); appDebug != "false" {
		bot.API.Debug = true
	}

	// Riporto a video lo stato di connessione
	log.Println("************************************************")
	log.Println("Bot connected: " + bot.API.Self.UserName)
	log.Println("************************************************")

	return
}

// GetUpdates - Ritorna nuovi messagi da lavorare da telegram
func (bot *Bot) GetUpdates() (tgbotapi.UpdatesChannel, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return bot.API.GetUpdatesChan(u)
}
