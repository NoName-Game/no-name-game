package config

import (
	"bitbucket.org/no-name-game/nn-telegram/init/connections/bot"
	"bitbucket.org/no-name-game/nn-telegram/init/connections/redis"
	"bitbucket.org/no-name-game/nn-telegram/init/connections/server"
	"bitbucket.org/no-name-game/nn-telegram/init/limiter"
	"bitbucket.org/no-name-game/nn-telegram/init/localization"
	"bitbucket.org/no-name-game/nn-telegram/init/logger"

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
		&App.Bot,
	} {
		service.Init()
	}

	App.Localization.Init(App.Server.Connection)
}

func init() {
	App = new(Configuration)
}
