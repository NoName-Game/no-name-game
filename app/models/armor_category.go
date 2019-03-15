package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// ArmorCategory -
type ArmorCategory struct {
	gorm.Model
	Slug string
	Name string
}

type ArmorCategories []ArmorCategory

// GetAllArmorCategories - Get all categories
func GetAllArmorCategories() (categories ArmorCategories) {
	services.Database.Find(&categories)

	return
}

// GetArmorCategoryBySlug - Get category by Slug
func GetArmorCategoryBySlug(slug string) (category ArmorCategory) {
	services.Database.Where("slug = ?", slug).First(&category)

	return
}

// SeederArmorCategory - Seeder Armor category
func SeederArmorCategory() {
	jsonFile, err := os.Open("resources/seeders/armor_categories.json")
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
		newArmorCategory := ArmorCategory{Name: category["name"], Slug: category["slug"]}
		services.Database.Where(ArmorCategory{Slug: category["slug"]}).FirstOrCreate(&newArmorCategory)
	}
}
