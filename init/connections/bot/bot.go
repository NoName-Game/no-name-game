package bot

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/joho/godotenv/autoload" // Autload .env
)

// Bot
type Bot struct {
	API *tgbotapi.BotAPI
}

// Init - Metodo per la connessione ai server telegram
func (bot *Bot) Init() {
	var err error

	// Recupero da env chiave telegram
	telegramAPIKey := os.Getenv("TELEGRAM_APIKEY")
	if telegramAPIKey == "" {
		logrus.WithField("error", fmt.Errorf("missing telegram apikey")).Fatal("[*] Telegram Connection: KO!")
	}

	// Istanzio comunicazione con il servizio dedicato
	if bot.API, err = tgbotapi.NewBotAPI(telegramAPIKey); err != nil {
		logrus.WithField("error", err).Fatal("[*] Telegram connection: KO!")
	}

	// Nel caso in cui fosse in ambiente di sviluppo abilito il debug
	if appDebug := os.Getenv("TELEGRAM_DEBUG"); appDebug != "false" {
		bot.API.Debug = true
	}

	logrus.WithField("Bot", bot.API.Self.UserName).Infof("[*] Telegram connection: OK!")
	return
}

// GetUpdates - Ritorna nuovi messagi da lavorare da telegram
func (bot *Bot) GetUpdates() (tgbotapi.UpdatesChannel, error) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return bot.API.GetUpdatesChan(u)
}
