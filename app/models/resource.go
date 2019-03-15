package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

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

// SeederResources - Seeder resources
func SeederResources() {
	jsonFile, err := os.Open("resources/seeders/resources.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		services.ErrorHandler("Error opening a file", err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var resources []map[string]string

	err = json.Unmarshal(byteValue, &resources)
	if err != nil {
		services.ErrorHandler("Error unmarshal resources seeder", err)
	}

	for _, resource := range resources {
		newResource := Resource{Name: resource["name"], Rarity: GetRarityBySlug(resource["rarity"]), ResourceCategory: GetCategoryBySlug(resource["category"])}

		services.Database.Where(Resource{Name: resource["name"]}).FirstOrCreate(&newResource)
	}
}
