package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Item - Item struct
type Item struct {
	gorm.Model
	Name           string `json:"name"`
	Rarity         Rarity
	RarityID       uint
	ItemCategory   ItemCategory
	ItemCategoryID uint
}

// Items - Items struct
type Items []Item

// GetItemByName - Get item by name
func GetItemByName(name string) Item {
	var item Item
	services.Database.Set("gorm:auto_preload", true).Where("name = ?", name).First(&item)

	return item
}

// GetItemByID - Get item by ID
func GetItemByID(id uint) Item {
	var item Item
	services.Database.Set("gorm:auto_preload", true).Where("id = ?", id).First(&item)

	return item
}

// SeederItems - Seeder items
func SeederItems() {
	jsonFile, err := os.Open("resources/seeders/items.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		services.ErrorHandler("Error opening a file", err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var items []map[string]string

	err = json.Unmarshal(byteValue, &items)
	if err != nil {
		services.ErrorHandler("Error unmarshal items seeder", err)
	}

	for _, item := range items {
		newItem := Item{Name: item["name"], Rarity: GetRarityBySlug(item["rarity"]), ItemCategory: GetCategoryBySlug(item["category"])}

		services.Database.Where(Item{Name: item["name"]}).FirstOrCreate(&newItem)
	}
}
