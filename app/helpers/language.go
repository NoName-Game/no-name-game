package helpers

import (
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/services"
)

// Trans - late shortCut
func Trans(key string, args ...interface{}) (message string) {
	// Getting default or player language
	locale := "en"
	var err error
	if Player.Language.Slug != "" {
		locale = Player.Language.Slug
	}

	// Check if the args are 0 or Null
	if len(args) <= 0 {
		message, err = services.GetTranslation(key, locale, nil)
	} else {
		message, err = services.GetTranslation(key, locale, args)
	}

	if err != nil {
		// If an error has been generated, it returns an empty string
		message = ""
	}

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

// GenerateTextArray - generate text's array from a common word in key.
func GenerateTextArray(common string) (texts []string) {
	var counter int
	for {
		keyText := common + "_" + strconv.Itoa(counter)
		if text := Trans(keyText); text != "" {
			texts = append(texts, text)
			counter++
		} else {
			break
		}
	}

	return
}
