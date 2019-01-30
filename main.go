package main

import (
	"gitlab.com/Valkyrie00/no-name/game"
	"gitlab.com/Valkyrie00/no-name/web"
)

func main() {
	// Server - Only for ping
	go web.Run()

	// Game - NoName
	game.Run()
}
