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
	migrations()
	seeders()

	services.RedisUp()

	//*************
	// Bot
	//*************
	services.BotUp()
}

func migrations() {
	services.Database.AutoMigrate(
		Player{},
		PlayerState{},
		Language{},
	)
}

func seeders() {
	seederLanguage()
}
