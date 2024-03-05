package localization

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"nn-grpc/build/pb"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

// Localization
type Localization struct {
	bundle *i18n.Bundle
}

// LanguageUp - Servizio di gestione multilingua
func (l *Localization) Init(serverConnection pb.NoNameClient) {
	var err error

	// Recupero lingue disponibili
	var languages []*pb.Language
	if languages, err = l.getAviableLanguages(serverConnection); err != nil {
		logrus.WithField("error", err).Fatal("[*] Languages: KO!")
	}

	// Creo bundle andando a caricare le diverse lingue
	if l.bundle, err = l.loadLocalizerBundle(languages); err != nil {
		logrus.WithField("error", err).Fatal("[*] Languages: KO!")
	}

	logrus.Info("[*] Languages: OK!")
	return
}

// getAviableLanguages - Recupero lingue disponibili
func (l *Localization) getAviableLanguages(serverConnection pb.NoNameClient) ([]*pb.Language, error) {
	var err error

	// Recupero tutte le lingue attive
	d := time.Now().Add(10 * time.Second)
	ctx, _ := context.WithDeadline(context.Background(), d)
	var rGetAllLanguages *pb.GetAllLanguagesResponse
	if rGetAllLanguages, err = serverConnection.GetAllLanguages(ctx, &pb.GetAllLanguagesRequest{}); err != nil {
		logrus.WithField("error", err).Fatal("[*] Languages: KO!")
	}

	return rGetAllLanguages.GetLanguages(), err
}

// CreateLocalizerBundle - Legge tutte le varie traduzione nei vari file e registra
func (l *Localization) loadLocalizerBundle(languages []*pb.Language) (bundle *i18n.Bundle, err error) {
	// Istanzio bundle con lingua di default
	bundle = i18n.NewBundle(language.English)

	// Abilito bundle a riconoscere i file di lingua
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// Ciclo traduzioni
	for _, lang := range languages {
		// Recupero solo le lingue attive
		if lang.GetEnabled() {
			// Recupero tutti i file di una specifica lingua
			files, err := ioutil.ReadDir(fmt.Sprintf("resources/lang/%s", lang.GetSlug()))
			if err != nil {
				return bundle, err
			}

			// Recupero tag della lingua
			tag := language.MustParse(lang.GetSlug())

			for _, file := range files {
				// Carico schema
				var message *i18n.MessageFile
				if message, err = bundle.LoadMessageFile(fmt.Sprintf("resources/lang/%s/%s", lang.GetSlug(), file.Name())); err != nil {
					return bundle, err
				}

				// Registro schema
				if err = bundle.AddMessages(tag, message.Messages...); err != nil {
					return bundle, err
				}
			}
		}
	}

	return
}

// GetTranslation - Return text from key translated to locale.
func (l *Localization) GetTranslation(key, locale string, args []interface{}) (string, error) {
	localizer := i18n.NewLocalizer(l.bundle, locale)
	msg, err := localizer.Localize(
		&i18n.LocalizeConfig{
			MessageID: key,
		},
	)

	return fmt.Sprintf(msg, args...), err
}
