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
		text, _ := services.GetTranslation("cron."+state.Function+"_alert", player.Language.Slug, nil)

		// Send notification
		msg := services.NewMessage(player.ChatID, text)
		// msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("continue", player.Language.Slug))))
		services.SendMessage(msg)

		// Update status
		state.ToNotify = false
		state.Update()
	}
}

// GetEndTime - Add to Now() the value passed.
func GetEndTime(hours, minutes, seconds int) (t time.Time) {
	t = time.Now().Add(time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds) + time.Second)
	return
}
