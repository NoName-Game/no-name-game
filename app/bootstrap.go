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

	//*************
	// Commands
	//*************
	// generate.Items()
	// generate.Stars()
}

func migrations() {
	services.Database.AutoMigrate(
		models.Star{},
		models.Player{},
		models.PlayerState{},
		models.PlayerStar{},
		models.PlayerPosition{},
		models.PlayerShip{},
		models.Language{},
		models.Rarity{},
		models.ItemCategory{},
		models.Item{},
		models.ShipCategory{},
		models.Ship{},
		models.Inventory{},
	)
}

func seeders() {
	models.SeederLanguage()
	models.SeederRarities()
	models.SeederItemCategory()
	models.SeederItems()
	models.SeederShipCategory()
}
