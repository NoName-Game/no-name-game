package helpers

import (
	"reflect"
	"strings"
	"time"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/app/provider"
	"bitbucket.org/no-name-game/no-name/services"
)

var (
	//===================================
	// Public

	Player nnsdk.Player

	//=====================================
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
	case reflect.Map:
		m := reflect.ValueOf(array)
		for _, e := range m.MapKeys() {
			if reflect.DeepEqual(val, m.MapIndex(e).Interface()) {
				exists = true
				return
			}
		}
	}

	return
}

// KeyInMap - Check if ID is in map
func KeyInMap(a uint, list map[uint]int) bool {
	for k := range list {
		if k == a {
			return true
		}
	}
	return false
}

// StringInSlice - Check if strin is in slice
func StringInSlice(v string, a []string) bool {
	for _, e := range a {
		if e == v {
			return true
		}
	}
	return false
}

// Slugger - convert text in slug
func Slugger(text string) string {
	//FIXME: replace me with reaplace all in Go 1.12
	return strings.Replace(strings.ToLower(text), " ", "_", -1)
}

// GetAllCategories - return all categories of all types
func GetAllCategories() (categories []string) {
	armorCategories, err := provider.GetAllArmorCategory()
	if err != nil {
		services.ErrorHandler("Cant get armor categories", err)
	}

	for _, armor := range armorCategories {
		categories = append(categories, armor.Name)
	}

	weaponCategories, err := provider.GetAllWeaponCategory()
	if err != nil {
		services.ErrorHandler("Cant get armor categories", err)
	}

	for _, weapon := range weaponCategories {
		categories = append(categories, weapon.Name)
	}

	return
}

// GetAllCategories - return all categories of all types
func GetAllSlugCategories() (categories []string) {
	armorCategories, err := provider.GetAllArmorCategory()
	if err != nil {
		services.ErrorHandler("Cant get armor categories", err)
	}

	for _, armor := range armorCategories {
		categories = append(categories, armor.Slug)
	}

	weaponCategories, err := provider.GetAllWeaponCategory()
	if err != nil {
		services.ErrorHandler("Cant get armor categories", err)
	}

	for _, weapon := range weaponCategories {
		categories = append(categories, weapon.Slug)
	}

	return
}

// GetEndTime - Add to Now() the value passed.
func GetEndTime(hours, minutes, seconds int) (t time.Time) {
	t = time.Now().Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds) + time.Second)
	return
}
