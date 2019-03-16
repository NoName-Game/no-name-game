package services

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	bundle *i18n.Bundle
	//Langs -
	Langs = map[string]string{
		"en": "English",
		"it": "Italian",
	}
)

// LanguageUp - LanguageUp
func LanguageUp() {
	var err error

	// Create a Bundle to use for the lifetime of your application
	bundle, err = createLocalizerBundle(Langs)
	if err != nil {
		ErrorHandler("Error initialising localization", err)
	}

	log.Println("************************************************")
	log.Println("Languages: OK!")
	log.Println("************************************************")
}

// CreateLocalizerBundle reads language files and registers them in i18n bundle
func createLocalizerBundle(Langs map[string]string) (*i18n.Bundle, error) {
	// Bundle stores a set of messages
	bundle = &i18n.Bundle{DefaultLanguage: language.Italian}

	// Enable bundle to understand yaml
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	var translations []byte
	var err error
	for file := range Langs {

		// Read our language yaml file
		translations, err = ioutil.ReadFile("resources/lang/" + file + ".yaml")
		if err != nil {
			ErrorHandler("Unable to read translation file", err)
			return nil, err
		}

		// It parses the bytes in buffer to add translations to the bundle
		bundle.MustParseMessageFileBytes(translations, "resources/lang/"+file+".yaml")
	}

	return bundle, nil
}

// GetTranslation - Return text from key translated to locale.
//
// You can use printf's placeholders!
// Available locales: it-IT, en-US
func GetTranslation(key, locale string, args ...interface{}) (string, error) {
	localizer := i18n.NewLocalizer(bundle, locale)
	msg, err := localizer.Localize(
		&i18n.LocalizeConfig{
			MessageID: key,
		},
	)

	msg = fmt.Sprintf(msg, args...)
	return msg, err
}
