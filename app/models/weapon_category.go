package models

import (
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
