package config

import (
	"errors"
	"os"

	raven "github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
)

const logFilePath = "storage/logs/log.json"

//Initialize the logger
func init() {
	if sentryDebug := os.Getenv("SENTRY_DEBUG"); sentryDebug != "false" {
		bootSentry()
	}

	if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
		bootLog()
	}
}

func bootSentry() {
	sentryDSN := os.Getenv("SENTRY_DSN")
	if sentryDSN == "" {
		ErrorHandler("$SENTRY_DSN must be set", errors.New("sentryDSN Missing"))
	}

	raven.SetDSN(sentryDSN)
}

func bootLog() {
	//Log file does not exist
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		_, err := os.Create(logFilePath)
		if err != nil {
			ErrorHandler("", err)
		}
	}
	file, _ := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, 0666)

	//DateFormatter
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	log.SetOutput(file)
}

// ErrorHandler logs a new Error.
//
// message is a custom text.
func ErrorHandler(message string, err error) {
	if err != nil {
		//Report to Sentry
		if sentryDebug := os.Getenv("SENTRY_DEBUG"); sentryDebug != "false" {
			raven.CaptureErrorAndWait(err, nil) //Invio errore potenzialmente Panicoso (Crusca fatti da parte) a Sentry
		}

		//Report Log
		if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
			log.WithFields(log.Fields{
				"Message": message,
			}).Panic(err)
		}
	}
}
