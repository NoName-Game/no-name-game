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
}

func migrations() {
	services.Database.AutoMigrate(
		models.Star{},
		models.Player{},
		models.PlayerState{},
		models.PlayerStar{},
		models.PlayerPosition{},
		models.Language{},
		models.Rarity{},
		models.Item{},
		models.Inventory{},
	)
}

func seeders() {
	models.SeederLanguage()
	models.SeederRarities()
	models.SeederItems()
}
