package game

import (
	"errors"

	"bitbucket.org/no-name-game/no-name/config"
)

func bootstrap() {
	config.ErrorHandler("Bootstrap", errors.New("Bootstrap"))
	migrations()
}

// Migrate the schema
func migrations() {
	config.Database.AutoMigrate(Player{})
}
