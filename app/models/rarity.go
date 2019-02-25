package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

type Rarities struct {
	gorm.Model
	Rarities []Rarity
}

type Rarity struct {
	gorm.Model
	Value string
	Slug  string
}

func GetRarityBySlug(slug string) Rarity {
	var rarity Rarity
	services.Database.Set("gorm:auto_preload", true).Where("value = ?", slug).First(&rarity)

	return rarity
}

func SeederRarities() {
	jsonFile, err := os.Open("resources/seeders/rarities.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		services.ErrorHandler("Error opening a file", err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var rarities Rarities

	json.Unmarshal(byteValue, &rarities)
	for _, rarity := range rarities.Rarities {
		newRarity := Rarity{Value: rarity.Value, Slug: rarity.Slug}
		services.Database.Where(Rarity{Value: rarity.Value}).FirstOrCreate(&newRarity)
	}
}
