package redis

import (
	"log"

	goredis "github.com/go-redis/redis"
)

// Redis
type Redis struct {
	Connection *goredis.Client
}

// RedisUp
func (redis *Redis) Init() {
	redis.Connection = goredis.NewClient(&goredis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if _, err := redis.Connection.Ping().Result(); err != nil {
		panic(err)
	}

	log.Println("************************************************")
	log.Println("Redis connection: OK!")
	log.Println("************************************************")

	return
}
