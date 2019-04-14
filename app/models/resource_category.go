package models

import (
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

