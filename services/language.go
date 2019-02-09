package services

import (
	"fmt"
	"io/ioutil"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	bundle    *i18n.Bundle
	langFiles = []string{
		"resources/lang/en-US/en-US.yaml",
		"resources/lang/it-IT/it-IT.yaml",
	}
)

// LanguageUp - LanguageUp
func LanguageUp() {
	var err error

	// Create a Bundle to use for the lifetime of your application
	bundle, err = createLocalizerBundle(langFiles)
	if err != nil {
		//ErrorHandling"Error initialising localization"
		panic(err)
	}
}

// CreateLocalizerBundle reads language files and registers them in i18n bundle
func createLocalizerBundle(langFiles []string) (*i18n.Bundle, error) {
	// Bundle stores a set of messages
	bundle := &i18n.Bundle{DefaultLanguage: language.Italian}

	// Enable bundle to understand yaml
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	var translations []byte
	var err error
	for _, file := range langFiles {

		// Read our language yaml file
		translations, err = ioutil.ReadFile(file)
		if err != nil {
			//ErrorHandling"Unable to read translation file"
			return nil, err
		}

		// It parses the bytes in buffer to add translations to the bundle
		bundle.MustParseMessageFileBytes(translations, file)
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
