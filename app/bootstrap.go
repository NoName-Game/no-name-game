package app

import (
	"bitbucket.org/no-name-game/nn-telegram/app/commands"
	"bitbucket.org/no-name-game/nn-telegram/services"
	_ "github.com/joho/godotenv/autoload" // Autload .env
)

// bootstrap - Carico servizi terzi
func bootstrap() (err error) {
	// *************
	// Logging - Servizi di logging errori
	// *************
	err = services.LoggingUp()
	if err != nil {
		return err
	}

	// *************
	// NoName WS - NoName Main server!
	// *************
	err = services.NnSDKUp()
	if err != nil {
		services.ErrorHandler("error connecting to NoName server", err)
		return err
	}

	// *************
	// Redis DB
	// *************
	err = services.RedisUp()
	if err != nil {
		services.ErrorHandler("error connecting to Redis", err)
		return err
	}

	// *************
	// i18n - Servizio di gestione multilingua
	// *************
	err = services.LanguageUp()
	if err != nil {
		services.ErrorHandler("error initialising localization", err)
		return err
	}

	// *************
	// Bot
	// *************
	err = services.BotUp()
	if err != nil {
		services.ErrorHandler("error booting bot", err)
		return err
	}

	// *************
	// Cron
	// *************
	// TODO: modificare
	go commands.Cron()

	return err
}
