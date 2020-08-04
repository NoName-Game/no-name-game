package commands

import (
	"os"
	"strconv"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

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

	for {
		// Sleep for minute
		time.Sleep(sleepTime)

		// Recupero tutto gli stati da notificare
		rGetPlayerStateToNotify, err := services.NnSDK.GetPlayerStateToNotify(helpers.NewContext(1), &pb.GetPlayerStateToNotifyRequest{})
		if err != nil {
			panic(err)
		}

		for _, state := range rGetPlayerStateToNotify.GetPlayerStates() {
			rGetPlayerByID, err := services.NnSDK.GetPlayerByID(helpers.NewContext(1), &pb.GetPlayerByIDRequest{
				ID: state.PlayerID,
			})
			if err != nil {
				panic(err)
			}

			// Recupero testo da notificare, ogni controller ha la propria notifica
			text, _ := services.GetTranslation("cron."+state.Controller+"_alert", rGetPlayerByID.GetPlayer().GetLanguage().GetSlug(), nil)

			// Invio notifica
			msg := services.NewMessage(rGetPlayerByID.GetPlayer().GetChatID(), text)
			// Al momento non associo nessun bottone potrebbe andare in conflitto con la mappa
			// continueButton, _ := services.GetTranslation(state.Function, player.Language.Slug, nil)
			// msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(continueButton)))
			_, err = services.SendMessage(msg)
			if err != nil {
				panic(err)
			}

			// Aggiorno lo stato levando la notifica
			state.ToNotify = false
			_, err = services.NnSDK.UpdatePlayerState(helpers.NewContext(1), &pb.UpdatePlayerStateRequest{
				PlayerState: state,
			})

			if err != nil {
				panic(err)
			}
		}
	}
}
