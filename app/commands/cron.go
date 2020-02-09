package commands

import (
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload" // Autload .env
)

// Cron - Call every minute the function
func Cron() {
	// TODO: da modificare ogni singolo cron dovrà avere il suo timeset, implemantere un'interfaccia

	envCronMinutes, _ := strconv.ParseInt(os.Getenv("CRON_MINUTES"), 36, 64)
	sleepTime := time.Duration(envCronMinutes) * time.Minute

	for {
		//Sleep for minute
		time.Sleep(sleepTime)

		//After sleep call function.
		CheckFinishTime()
	}
}
