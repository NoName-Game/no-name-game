package commands

import (
	"time"

	"bitbucket.org/no-name-game/no-name/app/provider"
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
	states, _ := provider.GetPlayerStateToNotify()

	for _, state := range states {
		player, _ := provider.GetPlayerByID(state.PlayerID)
		text, _ := services.GetTranslation("cron."+state.Function+"_alert", player.Language.Slug, nil)

		// Send notification
		msg := services.NewMessage(player.ChatID, text)
		services.SendMessage(msg)

		// Update status
		// Stupid poninter stupid json pff
		f := new(bool)
		*f = false

		state.ToNotify = f
		state, _ = provider.UpdatePlayerState(state)
	}
}

// GetEndTime - Add to Now() the value passed.
func GetEndTime(hours, minutes, seconds int) (t time.Time) {
	t = time.Now().Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds) + time.Second)
	return
}
