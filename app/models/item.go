package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"bitbucket.org/no-name-game/no-name/services"

	"github.com/jinzhu/gorm"
)

type Items struct {
	gorm.Model
	Items []Item `json:"items"`
}

type Item struct {
	gorm.Model
	Name   string `json:"name"`
	Rarity Rarity `json:"rarity"`
	// TODO: Add more information about item
}

func SeederItems() {
	jsonFile, err := os.Open("resources/seeders/items.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		services.ErrorHandler("Error opening a file", err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var items Items

	json.Unmarshal(byteValue, &items)
	for _, item := range items.Items {
		// FIXME: Non so al momento come inserire nel file json l'oggetto rarità
		newItem := Item{Name: item.Name, Rarity: GetRarityBySlug(item.Rarity.Slug)}
		services.Database.Where(Item{Name: item.Name}).FirstOrCreate(&newItem)
	}
}
