package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// ShipCategory -
type ShipCategory struct {
	gorm.Model
	Slug string
	Name string
}

type ShipCategories []ShipCategory

// GetAllShipCategories - Get all categories
func GetAllShipCategories() (categories ShipCategories) {
	services.Database.Find(&categories)

	return
}

// GetShipCategoryBySlug - Get category by Slug
func GetShipCategoryBySlug(slug string) (category ShipCategory) {
	services.Database.Where("slug = ?", slug).First(&category)

	return
}

// SeederShipCategory - Seeder Ship category
func SeederShipCategory() {
	jsonFile, err := os.Open("resources/seeders/ship_categories.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		services.ErrorHandler("Error opening a file", err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var categories []map[string]string

	err = json.Unmarshal(byteValue, &categories)
	if err != nil {
		services.ErrorHandler("Error unmarshal categories seeder", err)
	}

	for _, category := range categories {
		newShipCategory := ShipCategory{Name: category["name"], Slug: category["slug"]}
		services.Database.Where(ShipCategory{Slug: category["slug"]}).FirstOrCreate(&newShipCategory)
	}
}
