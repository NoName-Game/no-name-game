package main

import (
	"log"

	"gitlab.com/Valkyrie00/no-name/bot"
	"gitlab.com/Valkyrie00/no-name/config"
	"gitlab.com/Valkyrie00/no-name/web"
)

func main() {

	config.DatabaseConnection()

	// Server - Only for ping
	go web.Run()

	// Bot - NoName Game
	log.Println("LOAD - Bot")
	bot.Handler()
}
