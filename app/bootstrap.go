package app

import (
	"bitbucket.org/no-name-game/no-name/services"
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
	// Database
	//*************
	services.DatabaseUp()
	services.Database.AutoMigrate(Player{}, PlayerState{})

	//*************
	// Bot
	//*************
	services.BotUp()
}
