package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Language -
type Language struct {
	gorm.Model
	Slug  string
	Value string
}

// GetAllLangs - get all languages
func GetAllLangs() []Language {
	var languages []Language
	services.Database.Find(&languages)

	return languages
}

// GetLangByValue - get language by value
func GetLangByValue(lang string) Language {
	var language Language
	services.Database.Set("gorm:auto_preload", true).Where("value = ?", lang).First(&language)

	return language
}

// GetLangBySlug - get language by slug
func GetLangBySlug(lang string) Language {
	var language Language
	services.Database.Set("gorm:auto_preload", true).Where("slug = ?", lang).First(&language)

	return language
}

// SeederLanguage - SeederLanguage
func SeederLanguage() {
	for slug, lang := range services.Langs {
		newLanguage := Language{Value: lang, Slug: slug}
		services.Database.Where(Language{Slug: slug}).FirstOrCreate(&newLanguage)
	}

}
