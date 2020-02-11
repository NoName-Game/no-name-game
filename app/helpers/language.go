package helpers

import (
	"bitbucket.org/no-name-game/nn-telegram/services"
)

// Trans - late shortCut
func Trans(locale string, key string, args ...interface{}) (message string) {
	var err error
	if len(args) <= 0 {
		message, err = services.GetTranslation(key, locale, nil)
	} else {
		message, err = services.GetTranslation(key, locale, args)
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
// // GenerateTextArray - generate text's array from a common word in key.
// func GenerateTextArray(common string) (texts []string) {
// 	var counter int
// 	for {
// 		keyText := common + "_" + strconv.Itoa(counter)
// 		if text := Trans(keyText); text != "" {
// 			texts = append(texts, text)
// 			counter++
// 		} else {
// 			break
// 		}
// 	}
//
// 	return
// }
