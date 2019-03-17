package helpers

import (
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/no-name/app/models"
)

// InArray - check if val exist in array
func InArray(val interface{}, array interface{}) (exists bool) {
	exists = false

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				exists = true
				return
			}
		}
	}

	return
}

// StringInSlice
func KeyInMap(a string, list map[string]int) bool {
	for k := range list {
		if k == a {
			return true
		}
	}
	return false
}

// Slugger - convert text in slug
func Slugger(text string) string {
	return strings.ToLower(text)
}

// GetAllCategories - return all categories of all types
func GetAllCategories() (categories []string) {
	armorCategories := models.GetAllArmorCategories()
	for _, armor := range armorCategories {
		categories = append(categories, armor.Name)
	}

	weaponCategories := models.GetAllWeaponCategories()
	for _, weapon := range weaponCategories {
		categories = append(categories, weapon.Name)
	}

	return
}

// GetAllCategories - return all categories of all types
func GetAllSlugCategories() (categories []string) {
	armorCategories := models.GetAllArmorCategories()
	for _, armor := range armorCategories {
		categories = append(categories, armor.Slug)
	}

	weaponCategories := models.GetAllWeaponCategories()
	for _, weapon := range weaponCategories {
		categories = append(categories, weapon.Slug)
	}

	return
}
