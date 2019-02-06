package lang

import (
	"io/ioutil"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var bundle *i18n.Bundle

func init() {
	langFiles := []string{"en-US.yaml", "it-IT.yaml"}
	var err error

	// Create a Bundle to use for the lifetime of your application
	bundle, err = CreateLocalizerBundle(langFiles)
	if err != nil {
		//ErrorHandling"Error initialising localization"
		panic(err)
	}
}

// CreateLocalizerBundle reads language files and registers them in i18n bundle
func CreateLocalizerBundle(langFiles []string) (*i18n.Bundle, error) {
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

//Return text from key translated to locale.
//
//Available locales: it-IT, en-US
func getMessage(key, locale string) (string, error) {
	localizer := i18n.NewLocalizer(bundle, locale)
	msg, err := localizer.Localize(
		&i18n.LocalizeConfig{
			MessageID: key,
		},
	)
	return msg, err
}
