package app

import (
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
	// services.DatabaseUp()

	services.NnSDKUp()

	services.RedisUp()

	//*************
	// Bot
	//*************
	services.BotUp()

	// minutes, _ := strconv.ParseInt(os.Getenv("CRON_MINUTES"), 36, 64)
	// go commands.Cron(time.Duration(minutes) * time.Minute)

	//*************
	// Commands
	//*************
	// generate.Resources()
	// generate.Stars()
	// generate.Weapons()
	// generate.Armors()
	// generate.Ships()
	// generate.Enemies()
}
