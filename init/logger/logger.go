package logger

import (
	"log"
	"os"
	"time"

	"github.com/evalphobia/logrus_sentry"
	logrus "github.com/sirupsen/logrus"
)

// Logger
type Logger struct{}

// LoggingUp - Servizio per il caricamento dei vary servizi di logging
func (logger *Logger) Init() {
	enviroment := os.Getenv("ENV")

	// Message Formatter
	logrus.WithFields(logrus.Fields{
		"service": "nn-telegram-client",
		"env":     enviroment,
		"version": os.Getenv("version"),
	})
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Imposto livello di log
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)

	// Verifico integrazione con sentry
	if enviroment == "production" || enviroment == "staging" {
		// Escludo info message
		logrus.SetLevel(logrus.WarnLevel)

		if hook, err := logrus_sentry.NewSentryHook(os.Getenv("SENTRY_DSN"), []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		}); err == nil {
			hook.StacktraceConfiguration.Level = logrus.InfoLevel
			// hook.StacktraceConfiguration.Skip
			hook.StacktraceConfiguration.Context = 50
			// hook.Stacktrace  Configuration.InAppPrefixes
			hook.StacktraceConfiguration.IncludeErrorBreadcrumb = true
			hook.StacktraceConfiguration.Enable = true
			hook.Timeout = 10 * time.Second
			logrus.AddHook(hook)
		}
	}

	// Riporto a viodeo stato servizio
	log.Println("************************************************")
	log.Println("Logger: OK!")
	log.Println("************************************************")

	return
}
