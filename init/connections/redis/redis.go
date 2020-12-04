package redis

import (
	"fmt"
	"os"
	"strconv"

	goredis "github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

// Redis
type Redis struct {
	Connection *goredis.Client
}

// Redis Init
func (redis *Redis) Init() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		logrus.WithField("error", fmt.Errorf("missing redis host")).Fatal("[*] Redis Connection: KO!")
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		logrus.WithField("error", fmt.Errorf("missing redis port")).Fatal("[*] Redis Connection: KO!")
	}

	redisDB := os.Getenv("REDIS_DB")
	if redisDB == "" {
		logrus.WithField("error", fmt.Errorf("missing redis db")).Fatal("[*] Redis Connection: KO!")
	}

	// Converto indice DB
	redisDBIndex, _ := strconv.Atoi(redisDB)

	redis.Connection = goredis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: "",           // no password set
		DB:       redisDBIndex, // use default DB
	})

	if _, err := redis.Connection.Ping().Result(); err != nil {
		logrus.WithField("error", err).Fatal("[*] Redis: KO!")
	}

	logrus.Info("[*] Redis: OK!")
	return
}
