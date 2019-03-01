package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"bitbucket.org/no-name-game/no-name/services"
)

type Items struct {
	Items []Item `json:"items"`
}

type Item struct {
	ID          uint   `gorm:"primary_key"`
	Name        string `json:"name"`
	Rarity      Rarity `json:"-"`
	Rarity_slug string `json:"rarity"`
	// TODO: Add more information about item
}

func GetItemByName(name string) Item {
	var item Item
	services.Database.Set("gorm:auto_preload", true).Where("name = ?", name).First(&item)

	return item
}

func GetItemByID(id uint) Item {
	var item Item
	services.Database.Set("gorm:auto_preload", true).Where("id = ?", id).First(&item)

	return item
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
		newItem := Item{Name: item.Name, Rarity: GetRarityBySlug(item.Rarity_slug), Rarity_slug: item.Rarity_slug}
		services.Database.Where(Item{Name: item.Name}).FirstOrCreate(&newItem)
	}
}
