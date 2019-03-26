package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// WeaponCategory -
type WeaponCategory struct {
	gorm.Model
	Slug string
	Name string
}

type WeaponCategories []WeaponCategory

// GetAllWeaponCategories - Get all categories
func GetAllWeaponCategories() (categories WeaponCategories) {
	services.Database.Find(&categories)

	return
}

// GetWeaponCategoryBySlug - Get category by Slug
func GetWeaponCategoryBySlug(slug string) (category WeaponCategory) {
	services.Database.Where("slug = ?", slug).First(&category)

	return
}

// SeederWeaponCategory - Seeder Weapon category
func SeederWeaponCategory() {
	jsonFile, err := os.Open("resources/seeders/weapon_categories.json")
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
		newWeaponCategory := WeaponCategory{Name: category["name"], Slug: category["slug"]}
		services.Database.Where(WeaponCategory{Slug: category["slug"]}).FirstOrCreate(&newWeaponCategory)
	}
}
