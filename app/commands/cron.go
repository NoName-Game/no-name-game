package commands

import (
	"os"
	"strconv"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/services"
	_ "github.com/joho/godotenv/autoload" // Autload .env
)

// Cron
type Cron struct{}

// Notify - Metodo che si occupa di verificare e inviare le notifiche
func (c *Cron) Notify() {
	// Differisco controllo panic/recover
	defer func() {
		// Nel caso in cui panicasse
		if err := recover(); err != nil {
			// Registro errore
			services.ErrorHandler("cron recovered", err.(error))
		}
	}()

	// Recupero informazioni del cron
	envCronMinutes, _ := strconv.ParseInt(os.Getenv("CRON_MINUTES"), 36, 64)
	sleepTime := time.Duration(envCronMinutes) * time.Minute

	var playerStateProvicer providers.PlayerStateProvider
	var playerProvider providers.PlayerProvider

	for {
		// Sleep for minute
		time.Sleep(sleepTime)

		// Recupero tutto gli stati da notificare

		states, err := playerStateProvicer.GetPlayerStateToNotify()
		if err != nil {
			panic(err)
		}

		for _, state := range states {
			player, err := playerProvider.GetPlayerByID(state.PlayerID)
			if err != nil {
				panic(err)
			}

			// Recupero testo da notificare, ogni controller ha la propria notifica
			text, _ := services.GetTranslation("cron."+state.Controller+"_alert", player.Language.Slug, nil)

			// Invio notifica
			msg := services.NewMessage(player.ChatID, text)
			// Al momento non associo nessun bottone potrebbe andare in conflitto con la mappa
			// continueButton, _ := services.GetTranslation(state.Function, player.Language.Slug, nil)
			// msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(continueButton)))
			_, err = services.SendMessage(msg)
			if err != nil {
				panic(err)
			}

			// Aggiorno lo stato levando la notifica
			*state.ToNotify = false
			state, err = playerStateProvicer.UpdatePlayerState(state)
			if err != nil {
				panic(err)
			}
		}
	}
}
