package game

import (
	"bitbucket.org/no-name-game/no-name/config"
)

func bootstrap() {
	//*************
	// Errors Log
	//*************
	config.ErrorsUp()

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
	config.Database.AutoMigrate(Player{}, PlayerState{})
}
