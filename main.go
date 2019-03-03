package main

import (
	"bitbucket.org/no-name-game/no-name/app"
	"bitbucket.org/no-name-game/no-name/web"
)

func main() {
	// Server - Only for ping
	go web.Run()

	// Game - NoName
	app.Run()
}
