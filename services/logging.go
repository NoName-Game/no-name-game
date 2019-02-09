package services

import (
	"errors"
	"log"
	"os"

	raven "github.com/getsentry/raven-go"
	logrus "github.com/sirupsen/logrus"
)

const logFilePath = "storage/logs/errors.log"

//LoggingUp - LoggingUp
func LoggingUp() {
	if sentryDebug := os.Getenv("SENTRY_DEBUG"); sentryDebug != "false" {
		bootSentry()
	}

	if appDebug := os.Getenv("APP_DEBUG"); appDebug != "false" {
		bootLog()
	}

	log.Println("************************************************")
	log.Println("Errors Log: OK!")
	log.Println("************************************************")
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
	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		ErrorHandler("Error when opening file", err)
	}

	//DateFormatter
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	logrus.SetOutput(file)
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
			logrus.WithFields(logrus.Fields{
				"Message": message,
			}).Panic(err)
		}
	}
}
