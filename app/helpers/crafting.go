package helpers

import (
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
)

// ListRecipe - Metodo che aiuta a recuperare la lista di risore necessarie
// al crafting di un determianto item
func ListRecipe(needed map[uint]int) (result string, err error) {
	var resourceProvider providers.ResourceProvider
	for resourceID, value := range needed {
		var resource nnsdk.Resource
		resource, err = resourceProvider.GetResourceByID(resourceID)
		if err != nil {
			return result, err
		}

		result += resource.Name + " x" + strconv.Itoa(value) + "\n"
	}

	return result, err
}

func CheckAndReturnCategorySlug(locale string, text string) (result string) {
	for _, slug := range GetAllSlugCategories() {
		if text == Trans(locale, slug) {
			return slug
		}
	}
	return ""
}

func GetAllSlugCategories() (result []string) {
	var armorCategoriesProvider providers.ArmorCategoryProvider
	var weaponCategoriesProvider providers.WeaponCateogoryProvider
	aCategories, _ := armorCategoriesProvider.GetAllArmorCategory()
	wCategories, _ := weaponCategoriesProvider.GetAllWeaponCategory()
	for i := 0; i < len(aCategories); i++ {
		result = append(result, aCategories[i].Slug)
	}
	for i := 0; i < len(wCategories); i++ {
		result = append(result, wCategories[i].Slug)
	}
	return
}
