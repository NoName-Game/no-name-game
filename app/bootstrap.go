package app

import (
	"bitbucket.org/no-name-game/no-name/app/models"
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
		models.Player{},
		models.PlayerState{},
		models.Language{},
	)
}

func seeders() {
	models.SeederLanguage()
}
