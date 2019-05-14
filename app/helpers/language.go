package helpers

import (
	"bitbucket.org/no-name-game/no-name/services"
)

// Trans - late shortCut
func Trans(key string, args ...interface{}) (message string) {
	// Getting default or player language
	locale := "en"
	if Player.Language.Slug != "" {
		locale = Player.Language.Slug
	}

	if len(args) <= 0 {
		message, _ = services.GetTranslation(key, locale, nil)
		return
	}

	message, _ = services.GetTranslation(key, locale, args)
	return
}

// GetAllTranslatedSlugCategories - return weapon and armor slug category translated
func GetAllTranslatedSlugCategoriesByLocale() (results []string) {
	categories := GetAllSlugCategories()
	for _, category := range categories {
		results = append(results, Trans(category))
	}

	return
}
