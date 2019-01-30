package config

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/joho/godotenv/autoload"
)

var (
	// Database - Shared database connection
	Database *gorm.DB
)

// DatabaseConnection - Database Connection
func DatabaseConnection() {
	var err error

	connectionParameters := "host=" + os.Getenv("DB_HOST") + " port=" + os.Getenv("DB_PORT") + " user=" + os.Getenv("DB_USER") + " dbname=" + os.Getenv("DB_NAME") + " password=" + os.Getenv("DB_PASSWORD") + " sslmode=" + os.Getenv("DB_SSL")
	log.Println(connectionParameters)

	Database, err = gorm.Open("postgres", connectionParameters)
	if err != nil {
		log.Panicln(err)
	}
	defer Database.Close()

	log.Println("************************************************")
	log.Println("Connesso correttamente al DB.")
	log.Println("************************************************")
}
