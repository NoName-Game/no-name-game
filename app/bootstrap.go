package app

import (
	"os"
	"strconv"
	"time"

	"bitbucket.org/no-name-game/no-name/app/commands"
	"bitbucket.org/no-name-game/no-name/services"
	_ "github.com/joho/godotenv/autoload" // Autload .env
)

func bootstrap() {
	//*************
	// Logging
	//*************
	services.LoggingUp()

	//*************
	// i18n
	//*************
	services.LanguageUp()

	//*************
	// NoName WS
	//*************
	services.NnSDKUp()

	services.RedisUp()

	//*************
	// Bot
	//*************
	services.BotUp()

	minutes, _ := strconv.ParseInt(os.Getenv("CRON_MINUTES"), 36, 64)
	go commands.Cron(time.Duration(minutes) * time.Minute)
}
