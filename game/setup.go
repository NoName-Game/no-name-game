package game

import "bitbucket.org/no-name-game/no-name/config"

func bootstrap() {
	migrations()
}

// Migrate the schema
func migrations() {
	config.Database.AutoMigrate(Player{})
}
