package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Rarity - Rarity struct
type Rarity struct {
	gorm.Model
	Name string
	Slug string
}

// Rarities - Rarity slice
type Rarities []Rarity

// GetRarityBySlug - Get rarity by Slug
func GetRarityBySlug(slug string) Rarity {
	var rarity Rarity
	services.Database.Set("gorm:auto_preload", true).Where("slug = ?", slug).First(&rarity)

	return rarity
}

// SeederRarities - Seeder rarities
func SeederRarities() {
	jsonFile, err := os.Open("resources/seeders/rarities.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		services.ErrorHandler("Error opening a file", err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var rarities []map[string]string

	json.Unmarshal(byteValue, &rarities)
	for _, rarity := range rarities {
		newRarity := Rarity{Name: rarity["name"], Slug: rarity["slug"]}
		services.Database.Where(Rarity{Name: rarity["name"]}).FirstOrCreate(&newRarity)
	}
}
