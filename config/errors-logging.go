package config

import (
	"os"

	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

const filepath = "log.json"

//Inizialize the logger
func initialize() {
	raven.SetDSN(os.Getenv("SENTRY_DSN"))
	//Formato data
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	//Creo file log se non esiste
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		//does not exist
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
	//fmt.Println("Error Handling inizializzato")
}

// ErrorHandler logs a new Error.
//
// message is a custom text.
func ErrorHandler(message string, err error) {
	initialize()
	if err != nil {
		//Fatal Error
		raven.CaptureErrorAndWait(err, nil) //Invio errore potenzialmente Panicoso (Crusca fatti da parte) a Sentry
		log.WithFields(log.Fields{
			"Message": message,
		}).Panic(err) //Loggo su file
	}
}
