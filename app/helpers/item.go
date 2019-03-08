package helpers

import (
	"log"

	"bitbucket.org/no-name-game/no-name/app/commands/namer"
)

// TrainItemNames - Only for developer, this function generate a model for markov chain
// Call this function in game.go for example for generate modes.json
func TrainItemNames() {
	namer.TrainName("resources/items/", "names.txt")
}

// GenerateNewItems
func GenerateNewItems() {
	name := namer.GenerateName("resources/items/model.json")

	log.Panicln(name)
}
