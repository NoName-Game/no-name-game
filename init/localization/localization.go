package localization

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	// Langs - Lingue attualmente disponibili per questo client
	Langs = map[string]string{
		"en": "English",
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
		panic(err)
	}

	// Mostro a video stato servizio
	log.Println("************************************************")
	log.Println("Languages: OK!")
	log.Println("************************************************")

	return
}

// CreateLocalizerBundle - Legge tutte le varie traduzione nei vari file e registra
func (lang *Localization) loadLocalizerBundle() (bundle *i18n.Bundle, err error) {
	// Istanzio bundle con lingua di default
	bundle = i18n.NewBundle(language.English)

	// Abilito bundle a riconoscere i file di lingua
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// Ciclo traduzioni
	var translations []byte
	for file := range Langs {
		translations, err = ioutil.ReadFile("resources/lang/" + file + ".yaml")
		if err != nil {
			return bundle, err
		}

		// Parso il file
		bundle.MustParseMessageFileBytes(translations, "resources/lang/"+file+".yaml")
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
