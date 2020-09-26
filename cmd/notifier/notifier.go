package main

import (
	"os"
	"strconv"
	"time"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	"github.com/sirupsen/logrus"

	_ "github.com/joho/godotenv/autoload" // Autload .env
)

var (
	SleepTimer time.Duration
)

// Init
func init() {
	config.App.Bootstrap()

	// Recupero informazioni
	envMinutes, _ := strconv.ParseInt(os.Getenv("CRON_MINUTES"), 36, 64)
	SleepTimer = time.Duration(envMinutes) * time.Minute
}

func main() {
	for {
		var err error

		logrus.Info("[*] Start new loop...")

		// Recupero tutto gli stati da notificare
		var rGetPlayerActivityToNotify *pb.GetPlayerActivityToNotifyResponse
		if rGetPlayerActivityToNotify, err = config.App.Server.Connection.GetPlayerActivityToNotify(helpers.NewContext(1), &pb.GetPlayerActivityToNotifyRequest{}); err != nil {
			logrus.Panic(err)
		}

		logrus.Infof("[*] Notifications found: %d", len(rGetPlayerActivityToNotify.GetPlayerActivities()))
		for _, activity := range rGetPlayerActivityToNotify.GetPlayerActivities() {
			go handleNotification(activity)
		}

		// Sleep for minute
		time.Sleep(SleepTimer)
	}
}

func handleNotification(activity *pb.PlayerActivity) {
	var err error
	logrus.Infof("[*] Handle Activity: %d", activity.ID)

	defer func() {
		if err := recover(); err != nil {
			logrus.Info("[*] Activity %d recovered", activity.ID)
		}
	}()

	var rGetPlayerByID *pb.GetPlayerByIDResponse
	if rGetPlayerByID, err = config.App.Server.Connection.GetPlayerByID(helpers.NewContext(1), &pb.GetPlayerByIDRequest{
		ID: activity.PlayerID,
	}); err != nil {
		logrus.Panic(err)
	}

	// Recupero testo da notificare, ogni controller ha la propria notifica
	text, _ := config.App.Localization.GetTranslation("cron."+activity.Controller+"_alert", rGetPlayerByID.GetPlayer().GetLanguage().GetSlug(), nil)

	// Invio notifica
	msg := helpers.NewMessage(rGetPlayerByID.GetPlayer().GetChatID(), text)

	// Al momento non associo nessun bottone potrebbe andare in conflitto con la mappa
	// continueButton, _ := services.GetTranslation(state.Function, player.Language.Slug, nil)
	// msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(continueButton)))
	if _, err = helpers.SendMessage(msg); err != nil {
		logrus.Panic(err)
	}

	// Aggiorno lo stato levando la notifica
	if _, err = config.App.Server.Connection.SetPlayerActivityNotified(helpers.NewContext(1), &pb.SetPlayerActivityNotifiedRequest{
		ActivityID: activity.ID,
	}); err != nil {
		logrus.Panic(err)
	}
}
