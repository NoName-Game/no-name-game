package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Language -
type Language struct {
	gorm.Model
	Language string
}

// GetDefaultLangID - get Default Lang ID
func GetDefaultLangID(lang string) Language {
	var language Language
	services.Database.Set("gorm:auto_preload", true).Where("language = ?", lang).First(&language)

	return language
}

// SeederLanguage - SeederLanguage
func SeederLanguage() {
	for _, lang := range services.Langs {
		newLanguage := Language{Language: lang}
		services.Database.Where(Language{Language: lang}).FirstOrCreate(&newLanguage)
	}
}
