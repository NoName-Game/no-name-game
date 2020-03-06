package services

import (
	"errors"
	"log"
	"os"
	"time"

	sentry "github.com/getsentry/sentry-go"

	logrus "github.com/sirupsen/logrus"
)

// Path di dove salvare il log degli errori
const logFilePath = "storage/logs/errors.log"

// LoggingUp - Servizio per il caricamento dei vary servizi di logging
func LoggingUp() (err error) {
	// Verifico se a livello di env è abilitato sentry
	if sentryDebug := os.Getenv("SENTRY_DEBUG"); sentryDebug != "false" {
		// Istanzio connessione a Sentry
		err = bootSentry()
		if err != nil {
			return err
		}
	}

	// Imposto settaggi di formattazione per scrittura su file
	err = bootLog()
	if err != nil {
		return err
	}

	// Riporto a viodeo stato servizio
	log.Println("************************************************")
	log.Println("Errors Log: OK!")
	log.Println("************************************************")

	return
}

// Motodo di bot del servizio sentry
func bootSentry() (err error) {
	// Recupero sentry dsn
	var dsn = os.Getenv("SENTRY_DSN")
	if dsn == "" {
		err = errors.New("missing ENV sentryDSN")
		return err
	}

	// Recupero enviroment di lavoro per sentry | opzionale
	var enviroment string
	enviroment = os.Getenv("SENTRY_ENV")
	if enviroment == "" {
		enviroment = "dev"
	}

	// Inizializzo connessione a sentry
	err = sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: enviroment,
		Debug:       true,
		DebugWriter: os.Stderr,
	})

	return err
}

// Metodo per il settaggio del serivizo di loggin su file
func bootLog() (err error) {
	// Verifico se esiste il file di logging
	var file *os.File
	file, err = os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return err
	}

	// DateFormatter
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Imposto livello di loggin
	logrus.SetLevel(logrus.InfoLevel)

	// Imposto file output
	logrus.SetOutput(file)

	return
}

// ErrorHandler - Metodo per la gestione e distrubuzione degli errori
func ErrorHandler(message string, err error) {
	// Se sono in fase di sviluppo non registro niente e riporto l'errore
	// if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
	// 	log.Panicln(err)
	// }

	// Registro errore su sentry
	if sentryDebug := os.Getenv("SENTRY_DEBUG"); sentryDebug != "false" {
		sentry.CaptureException(err)
		_ = sentry.Flush(time.Second)
	}

	// Registro errore in log file
	logrus.WithFields(logrus.Fields{
		"Message": message,
	}).Error(err)
}
