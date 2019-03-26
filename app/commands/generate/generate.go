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

// Resources - this command generate a new list of items for seeder.
func Resources() {
	// Step 1 - Generate makov chain model
	// namer.TrainName("resources/namer/items/", "names.txt")

	// Step 2 - Generate seeder
	type SeederResource map[string]string
	type SeederResources []SeederResource

	var items SeederResources

	// type jsonStruct []map[string]string
	for _, category := range models.GetAllResourceCategories() {
		for _, rarity := range models.GetAllRarities() {
			for i := 1; i <= 5; i++ {
				name := namer.GenerateName("resources/namer/items/model.json")
				item := SeederResource{
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
// (it's used only for testing galaxy structure)
func Stars() {
	// Step 1 - Generate makov chain model
	// namer.TrainName("resources/namer/stars/", "names.txt")

	// Step 2 - Generate galaxy
	nUsers := 10
	for deviation := 1; deviation <= nUsers; deviation++ {
		helpers.NewGalaxyChunk(deviation)
	}

	log.Println("************************************************")
	log.Println("End star generator")
	log.Println("************************************************")
}

// Ship - this command generate a new ship.
// (it's used only for testing ship creation)
func Ships() {
	// Step 1 - Generate makov chain model
	// namer.TrainName("resources/namer/ships/", "names.txt")

	// Step 2 - Generate ship
	helpers.NewStartShip()

	log.Println("************************************************")
	log.Println("End Ship generator")
	log.Println("************************************************")
}

// Weapon - this command generate a new weapon.
// (it's used only for testing weapon creation)
func Weapons() {
	// Step 1 - Generate makov chain model
	// namer.TrainName("resources/namer/weapons/", "names.txt")

	// Step 2 - Generate ship
	helpers.NewWeapon()

	log.Println("************************************************")
	log.Println("End Weapon generator")
	log.Println("************************************************")
}

// Armors - this command generate a new armor.
// (it's used only for testing armor creation)
func Armors() {
	// Step 1 - Generate makov chain model
	// namer.TrainName("resources/namer/armors/", "names.txt")

	// Step 2 - Generate ship
	helpers.NewArmor()

	log.Println("************************************************")
	log.Println("End Armor generator")
	log.Println("************************************************")
}
