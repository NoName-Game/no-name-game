package localization

import (
	"fmt"
	"io/ioutil"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	// Langs - Lingue attualmente disponibili per questo client
	Langs = map[string]string{
		// "en": "English",
		"it": "Italian",
	}
)

// Localization
type Localization struct {
	bundle *i18n.Bundle
}

// LanguageUp - Servizio di gestione multilingua
func (lang *Localization) Init() {
	var err error

	// Creo bundle andando a caricare le diverse lingue
	if lang.bundle, err = lang.loadLocalizerBundle(); err != nil {
		logrus.WithField("error", err).Fatal("[*] Languages: KO!")
	}

	logrus.Info("[*] Languages: OK!")
	return
}

// CreateLocalizerBundle - Legge tutte le varie traduzione nei vari file e registra
func (lang *Localization) loadLocalizerBundle() (bundle *i18n.Bundle, err error) {
	// Istanzio bundle con lingua di default
	bundle = i18n.NewBundle(language.English)

	// Abilito bundle a riconoscere i file di lingua
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// Ciclo traduzioni
	for lang := range Langs {
		// Recupero tutti i file di una specifica lingua
		files, err := ioutil.ReadDir(fmt.Sprintf("resources/lang/%s", lang))
		if err != nil {
			return bundle, err
		}

		for _, file := range files {
			// Carico schema
			var message *i18n.MessageFile
			if message, err = bundle.LoadMessageFile(fmt.Sprintf("resources/lang/%s/%s", lang, file.Name())); err != nil {
				return bundle, err
			}

			// Registro schema
			if err = bundle.AddMessages(language.English, message.Messages...); err != nil {
				return bundle, err
			}
		}
	}

	return
}

// GetTranslation - Return text from key translated to locale.
//
// You can use printf's placeholders!
// Available locales: it-IT, en-US
func (lang *Localization) GetTranslation(key, locale string, args []interface{}) (string, error) {
	localizer := i18n.NewLocalizer(lang.bundle, locale)
	msg, err := localizer.Localize(
		&i18n.LocalizeConfig{
			MessageID: key,
		},
	)

	msg = fmt.Sprintf(msg, args...)
	return msg, err
}
