package services

import (
	"log"

	"github.com/go-redis/redis"
)

var (
	// Redis - Shared redis connection
	Redis *redis.Client
)

// RedisUp
func RedisUp() (err error) {
	Redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err = Redis.Ping().Result()

	log.Println("************************************************")
	log.Println("Redis connection: OK!")
	log.Println("************************************************")

	return
}
