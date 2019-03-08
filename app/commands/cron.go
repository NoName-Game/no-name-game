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
	for _, state := range models.GetAllPlayerState() {
		if time.Now().Before(state.FinishAt) && state.DeletedAt == nil && state.ToNotify {
			player := state.Player
			text, _ := services.GetTranslation(state.Function, player.Language.Slug)
			services.SendMessage(services.NewMessage(player.ChatID, text))
			state.ToNotify = false
			state.Update()
		}
	}
}
