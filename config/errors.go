package config

import (
	"errors"
	"os"

	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

const filepath = "log.json"

//Initialize the logger
func init() {
	sentryDSN := os.Getenv("SENTRY_DSN")
	if sentryDSN == "" {
		ErrorHandler("$SENTRY_DSN must be set", errors.New("sentryDSN Missing"))
	}

	raven.SetDSN(sentryDSN)

	//DateFormatter
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		//File does not exist
		file, err := os.Create(filepath)
		log.SetOutput(file)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
		}
	} else {
		file, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
		}
		log.SetOutput(file)
	}
}

// ErrorHandler logs a new Error.
//
// message is a custom text.
func ErrorHandler(message string, err error) {
	if err != nil {
		//Report to Sentry
		raven.CaptureErrorAndWait(err, nil) //Invio errore potenzialmente Panicoso (Crusca fatti da parte) a Sentry

		//Log file
		log.WithFields(log.Fields{
			"Message": message,
		}).Panic(err)
	}
}
