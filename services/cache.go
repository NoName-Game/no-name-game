package services

import (
	"log"
	"time"

	_ "github.com/joho/godotenv/autoload" // Autoload .env
	gocache "github.com/patrickmn/go-cache"
)

var (
	// Cache
	Cache *gocache.Cache
)

// CacheUp - Starte GoCache
func CacheUp() (err error) {
	Cache = gocache.New(10*time.Minute, 10*time.Minute)

	// Riporto a video stato servizio
	log.Println("************************************************")
	log.Println("Cache Memory: OK!")
	log.Println("************************************************")

	return
}
