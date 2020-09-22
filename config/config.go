package config

import (
	"bitbucket.org/no-name-game/nn-telegram/init/connections/bot"
	"bitbucket.org/no-name-game/nn-telegram/init/connections/redis"
	"bitbucket.org/no-name-game/nn-telegram/init/connections/server"

	_ "github.com/joho/godotenv/autoload" // Autload .env
)

var App *Configuration

type Configuration struct {
	Bot    bot.Bot
	Redis  redis.Redis
	Server server.Server
}

func init() {
	App = new(Configuration)
}
