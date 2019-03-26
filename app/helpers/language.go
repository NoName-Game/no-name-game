package helpers

import (
	"bitbucket.org/no-name-game/no-name/services"
)

// Trans - late shortCut
func Trans(key, locale string, args ...interface{}) (message string) {
	if len(args) <= 0 {
		message, _ = services.GetTranslation(key, locale, nil)
		return
	}

	message, _ = services.GetTranslation(key, locale, args)
	return
}

// GetAllTranslatedSlugCategories - return weapon and armor slug category translated
func GetAllTranslatedSlugCategoriesByLocale(locale string) (results []string) {
	categories := GetAllSlugCategories()
	for _, category := range categories {
		results = append(results, Trans(category, locale))
	}

	return
}
