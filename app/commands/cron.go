package commands

import (
	"time"

	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
)

// Cron - Call every minute the function
func Cron(minute time.Duration) {
	for {
		//Sleep for minute
		time.Sleep(minute)
		//After sleep call function.
		CheckFinishTime()
	}
}

// CheckFinishTime - Check the ending and handle the functions.
func CheckFinishTime() {
	for _, state := range models.GetAllStateToNotify() {
		player := models.FindPlayerByID(state.PlayerID)
		text, _ := services.GetTranslation(state.Function, player.Language.Slug)

		// Send notification
		services.SendMessage(services.NewMessage(player.ChatID, text))

		// Update status
		state.ToNotify = false
		state.Update()
	}
}
