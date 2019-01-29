// main.go
package main

import (
	"log"

	"gitlab.com/Valkyrie00/no-name/bot"
)

func main() {
	// PingServer
	log.Println("LOAD - Ping Server")
	go server()

	log.Println("LOAD - Bot")
	bot.Handler()
}
