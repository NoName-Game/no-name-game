package bot

import (
	"fmt"
	"log"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/joho/godotenv/autoload"
	"gitlab.com/Valkyrie00/no-name/config"
)

var (
	bot *tgbotapi.BotAPI
)

func init() {
	telegramAPIKey := os.Getenv("TELEGRAM_APIKEY")
	if telegramAPIKey == "" {
		panic("$TELEGRAM_APIKEY must be set")
	}

	var botErr error
	bot, botErr = tgbotapi.NewBotAPI(telegramAPIKey)
	bot.Debug = true

	if botErr != nil {
		log.Panic(botErr)
	}

	log.Println(fmt.Sprintf("Bot connected: %s", bot.Self.UserName))

	// Da mettere in game fare migrations
	migrations()
}

// Migrate the schema
func migrations() {
	config.Database.AutoMigrate(User{})
}

//Handler - Updates Handler
func Handler() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}
