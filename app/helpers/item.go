package helpers

import (
	"encoding/json"
	"log"

	"bitbucket.org/no-name-game/no-name/app/commands/namer"
	"bitbucket.org/no-name-game/no-name/app/models"
)

// TrainItemNames - Only for developer, this function generate a model for markov chain
// Call this function in game.go for example for generate modes.json
func TrainItemNames() {
	namer.TrainName("resources/items/", "names.txt")
}

// GenerateNewItems
func GenerateNewItems() {
	var items models.Items
	type jsonStruct []map[string]string
	for _, category := range models.GetAllItemCategories() {
		for _, rarity := range models.GetAllRarities() {
			for i := 0; i < 5; i++ {
				name := namer.GenerateName("resources/items/model.json")
				newItem := jsonStruct{
					"name":     name,
					"rarity":   rarity.Slug,
					"category": category,
				}
				//TODO: fix me
				// items = append(items, newItem)
			}
		}

	}

	jsonObj, _ := json.Marshal(items)
	log.Panicln(string(jsonObj))
}
