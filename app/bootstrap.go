package app

import (
	"os"
	"strconv"
	"time"

	"bitbucket.org/no-name-game/no-name/app/commands"
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

	minutes, _ := strconv.ParseInt(os.Getenv("CRON_MINUTES"), 36, 64)
	go commands.Cron(time.Duration(minutes) * time.Minute)

	//*************
	// Commands
	//*************
	// generate.Resources()
	// generate.Stars()
	// generate.Weapons()
	// generate.Armors()
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
		models.ResourceCategory{},
		models.Resource{},
		models.ShipCategory{},
		models.Ship{},
		models.WeaponCategory{},
		models.Weapon{},
		models.ArmorCategory{},
		models.Armor{},
		models.Inventory{},
	)
}

func seeders() {
	models.SeederLanguage()
	models.SeederRarities()
	models.SeederResourceCategory()
	models.SeederResources()
	models.SeederShipCategory()
	models.SeederWeaponCategory()
	models.SeederArmorCategory()
}
