package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// ItemCategory -
type ItemCategory struct {
	gorm.Model
	Slug string
	Name string
}

type ItemCategories []ItemCategory

// GetAllItemCategories - Get all categories
func GetAllItemCategories() ItemCategories {
	var categories ItemCategories
	services.Database.Find(&categories)

	return categories
}

// GetCategoryBySlug - Get category by Slug
func GetCategoryBySlug(slug string) ItemCategory {
	var category ItemCategory
	services.Database.Where("slug = ?", slug).First(&category)

	return category
}

// SeederItemCategory - Seeder item category
func SeederItemCategory() {
	jsonFile, err := os.Open("resources/seeders/item_categories.json")
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
		newItemCategory := ItemCategory{Name: category["name"], Slug: category["slug"]}
		services.Database.Where(ItemCategory{Slug: category["slug"]}).FirstOrCreate(&newItemCategory)
	}
}
