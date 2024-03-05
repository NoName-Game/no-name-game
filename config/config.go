package config

import (
	"nn-telegram/init/antiflood"
	"nn-telegram/init/connections/bot"
	"nn-telegram/init/connections/redis"
	"nn-telegram/init/connections/server"
	"nn-telegram/init/limiter"
	"nn-telegram/init/localization"
	"nn-telegram/init/logger"

	_ "github.com/joho/godotenv/autoload" // Autload .env
)

var (
	App *Configuration
)

type Configuration struct {
	Bot          bot.Bot
	Logger       logger.Logger
	Redis        redis.Redis
	Server       server.Server
	Localization localization.Localization
	RateLimiter  limiter.RateLimiter
	Antiflood    antiflood.Antiflood
}

type GameService interface {
	Init()
}

// Bootstrap - Inizializzo servizi
func (config *Configuration) Bootstrap() {
	for _, service := range []GameService{
		&App.Logger,
		&App.Redis,
		&App.Server,
		&App.RateLimiter,
		&App.Antiflood,
		&App.Bot,
	} {
		service.Init()
	}

	App.Localization.Init(App.Server.Connection)
}

func init() {
	App = new(Configuration)
}
