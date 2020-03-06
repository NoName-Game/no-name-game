package services

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
	_ "github.com/joho/godotenv/autoload" // Autoload .env
)

var (
	// Redis - Shared redis connection
	Redis *redis.Client
)

// RedisUp - Inizializzo comunicazioe a RedisDB
func RedisUp() (err error) {
	// Recupero Host
	var host = os.Getenv("REDIS_HOST")
	if host == "" {
		err = errors.New("missing ENV REDIS_HOST")
		return err
	}

	// Recupero porta
	var port = os.Getenv("REDIS_PORT")
	if port == "" {
		err = errors.New("missing ENV REDIS_PORT")
		return err
	}

	// Inizializzo connessione a redis
	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})

	// Riporto a video stato servizio
	log.Println("************************************************")
	log.Println("Redis connection: OK!")
	log.Println("************************************************")

	return
}
