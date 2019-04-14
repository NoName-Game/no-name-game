package models

import (
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
