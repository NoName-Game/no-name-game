package main

import "log"

var (
	starNames = []string{
		"Acamar", "Achernar", "Achird", "Acrab",
	}
)

func main() {
	// // Server - Only for ping
	// go web.Run()

	// // Game - NoName
	// app.Run()

	generate()
}

func generate() {
	log.Println("hello")
}
