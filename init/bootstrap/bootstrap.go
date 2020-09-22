package bootstrap

import (
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/init/languages"
	"bitbucket.org/no-name-game/nn-telegram/init/logging"
	_ "github.com/joho/godotenv/autoload" // Autload .env
)

// bootstrap - Carico servizi terzi
func Bootstrap() (err error) {
	// *************
	// Logging - Servizi di logging errori
	// *************
	err = logging.LoggingUp()
	if err != nil {
		return err
	}

	// *************
	// NoName WS - NoName Main server!
	// *************
	config.App.Server.Init()

	// *************
	// Cache
	// *************
	config.App.Redis.Init()
	// err = services.CacheUp()
	// err = redis.RedisUp()
	// if err != nil {
	// 	logging.ErrorHandler("error starting cache", err)
	// 	return err
	// }

	// *************
	// i18n - Servizio di gestione multilingua
	// *************
	err = languages.LanguageUp()
	if err != nil {
		logging.ErrorHandler("error initialising localization", err)
		return err
	}

	// *************
	// Bot
	// *************
	config.App.Bot.Init()
	// err = bot.BotUp()
	// if err != nil {
	// 	logging.ErrorHandler("error booting bot", err)
	// 	return err
	// }

	// *************
	// Cron
	// *************
	// var cron commands.Cron
	// go cron.Notify()

	return err
}
