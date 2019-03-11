package generate

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"bitbucket.org/no-name-game/no-name/app/acme/namer"
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
)

// Items - this command generate a new list of items for seeder.
func Items() {
	// Step 1 - Generate makov chain model
	// namer.TrainName("resources/items/", "names.txt")

	// Step 2 - Generate seeder
	type SeederItem map[string]string
	type SeederItems []SeederItem

	var items SeederItems

	// type jsonStruct []map[string]string
	for _, category := range models.GetAllItemCategories() {
		for _, rarity := range models.GetAllRarities() {
			for i := 1; i <= 5; i++ {
				name := namer.GenerateName("resources/items/model.json")
				item := SeederItem{
					"name":     name,
					"rarity":   rarity.Slug,
					"category": category.Slug,
				}
				items = append(items, item)
			}
		}

	}

	jsonObj, _ := json.Marshal(items)
	err := ioutil.WriteFile("resources/seeders/items.json", jsonObj, 0644)
	if err != nil {
		services.ErrorHandler("Error DEL player state in redis", err)
	}

	log.Println("************************************************")
	log.Println("End items generator")
	log.Println("************************************************")
}

// Stars - this command generate a new galaxy.
// (it's used only for testing falaxy structure)
func Stars() {
	// Step 1 - Generate makov chain model
	// namer.TrainName("resources/stars/", "names.txt")

	// Step 2 - Generate galaxy
	nUsers := 10
	for deviation := 1; deviation <= nUsers; deviation++ {
		helpers.NewGalaxyChunk(deviation)
	}

	log.Println("************************************************")
	log.Println("End star generator")
	log.Println("************************************************")
}
