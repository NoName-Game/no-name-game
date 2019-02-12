package services

import (
	"log"

	"github.com/go-redis/redis"
	_ "github.com/joho/godotenv/autoload"
)

var (
	// Redis - Shared redis connection
	Redis *redis.Client
)

//RedisUp - RedisUp
func RedisUp() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	log.Println("************************************************")
	log.Println("Redis connection: OK!")
	log.Println("************************************************")
}
