package helpers

import (
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/init/languages"
)

// Trans - late shortCut
func Trans(locale string, key string, args ...interface{}) (message string) {
	var err error
	if len(args) <= 0 {
		message, err = languages.GetTranslation(key, locale, nil)
	} else {
		message, err = languages.GetTranslation(key, locale, args)
	}

	if err != nil {
		panic(err)
	}

	return
}

// GetAllTranslatedSlugCategories - return weapon and armor slug category translated
// func GetAllTranslatedSlugCategoriesByLocale() (results []string) {
// 	categories := GetAllSlugCategories()
// 	for _, category := range categories {
// 		results = append(results, Trans(category))
// 	}
//
// 	return
// }
//

// GenerateTextArray - Recupero un serie di testi definiti dalla stessa chiave e numerati
func GenerateTextArray(locale string, common string) (texts []string) {
	var counter int

	for {
		keyText := common + "_" + strconv.Itoa(counter)

		var translatedText string
		translatedText, _ = languages.GetTranslation(keyText, locale, nil)

		if translatedText != "" {
			texts = append(texts, translatedText)
			counter++
		} else {
			break
		}
	}

	return
}
