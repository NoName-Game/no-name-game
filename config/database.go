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

// Init Database Connection
func init() {
	var err error
	connectionParameters := "host=" + os.Getenv("DB_HOST") + " port=" + os.Getenv("DB_PORT") + " user=" + os.Getenv("DB_USER") + " dbname=" + os.Getenv("DB_NAME") + " password=" + os.Getenv("DB_PASSWORD") + " sslmode=" + os.Getenv("DB_SSL")
	Database, err = gorm.Open("postgres", connectionParameters)
	if err != nil {
		log.Panicln(err)
	}

	log.Println("************************************************")
	log.Println("Database connection: OK!")
	log.Println("************************************************")
}
