package commands

import (
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/no-name-game/no-name/app/provider"
	"bitbucket.org/no-name-game/no-name/services"
	_ "github.com/joho/godotenv/autoload" // Autload .env
)

// Cron - Call every minute the function
func Cron() {
	envCronMinutes, _ := strconv.ParseInt(os.Getenv("CRON_MINUTES"), 36, 64)
	sleepTime := time.Duration(envCronMinutes) * time.Minute

	for {
		//Sleep for minute
		time.Sleep(sleepTime)

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
		continueButton, _ := services.GetTranslation(state.Function, player.Language.Slug, nil)
		// I need this continue button to recall the function.
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(continueButton)))
		services.SendMessage(msg)

		// Update status
		// Stupid poninter stupid json pff
		f := new(bool)
		*f = false

		state.ToNotify = f
		state, _ = provider.UpdatePlayerState(state)
	}
}
