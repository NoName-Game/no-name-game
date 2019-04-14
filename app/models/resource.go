package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Resource - Resource struct
type Resource struct {
	gorm.Model
	Name               string `json:"name"`
	Rarity             Rarity
	RarityID           uint
	ResourceCategory   ResourceCategory
	ResourceCategoryID uint
}

// Resources - Resources struct
type Resources []Resource

// GetResourceByName - Get Resource by name
func GetResourceByName(name string) Resource {
	var resource Resource
	services.Database.Set("gorm:auto_preload", true).Where("name = ?", name).First(&resource)

	return resource
}

// GetResourceByID - Get resource by ID
func GetResourceByID(id uint) Resource {
	var resource Resource
	services.Database.Set("gorm:auto_preload", true).Where("id = ?", id).First(&resource)

	return resource
}

// GetRandomResourceByCategory
func GetRandomResourceByCategory(categoryID uint) Resource {
	var resource Resource
	services.Database.Set("gorm:auto_preload", true).Where("resource_category_id = ?", categoryID).Order(gorm.Expr("random()")).First(&resource)

	return resource
}
