package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// ResourceCategory -
type ResourceCategory struct {
	gorm.Model
	Slug string
	Name string
}

type ResourceCategories []ResourceCategory

// GetAllResourceCategories - Get all categories
func GetAllResourceCategories() ResourceCategories {
	var categories ResourceCategories
	services.Database.Find(&categories)

	return categories
}

// GetCategoryBySlug - Get category by Slug
func GetCategoryBySlug(slug string) ResourceCategory {
	var category ResourceCategory
	services.Database.Where("slug = ?", slug).First(&category)

	return category
}

// SeederResourceCategory - Seeder Resource category
func SeederResourceCategory() {
	jsonFile, err := os.Open("resources/seeders/resource_categories.json")
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
		newResourceCategory := ResourceCategory{Name: category["name"], Slug: category["slug"]}
		services.Database.Where(ResourceCategory{Slug: category["slug"]}).FirstOrCreate(&newResourceCategory)
	}
}
