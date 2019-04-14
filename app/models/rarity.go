package models

import (
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

// GetAllCategories - Get all rarities
func GetAllRarities() Rarities {
	var rarities Rarities
	services.Database.Find(&rarities)

	return rarities
}

// GetRarityBySlug - Get rarity by Slug
func GetRarityBySlug(slug string) Rarity {
	var rarity Rarity
	services.Database.Set("gorm:auto_preload", true).Where("slug = ?", slug).First(&rarity)

	return rarity
}
