package game

import (
	"log"

	"bitbucket.org/no-name-game/no-name/config"
)

func bootstrap() {
	//*************
	// Errors Log
	//*************
	config.ErrorsUp()

	//*************
	// i18n
	//*************
	config.LanguageUp()

	//*************
	// Database
	//*************
	config.DatabaseUp()
	migrations()

	//*************
	// Bot
	//*************
	config.BotUp()
}

// Migrate the schema
func migrations() {
	log.Println("Migrations")
	config.Database.AutoMigrate(Player{}, PlayerState{})
}
