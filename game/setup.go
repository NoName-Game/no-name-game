package game

import "gitlab.com/Valkyrie00/no-name/config"

func bootstrap() {
	migrations()
}

// Migrate the schema
func migrations() {
	config.Database.AutoMigrate(Player{}, PlayerState{})
}
