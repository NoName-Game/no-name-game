package redis

import (
	goredis "github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

// Redis
type Redis struct {
	Connection *goredis.Client
}

// Redis Init
func (redis *Redis) Init() {
	redis.Connection = goredis.NewClient(&goredis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if _, err := redis.Connection.Ping().Result(); err != nil {
		logrus.WithField("error", err).Fatal("[*] Redis: KO!")
	}

	logrus.Info("[*] Redis: OK!")
	return
}
