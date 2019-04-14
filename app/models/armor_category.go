package models

import (
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
