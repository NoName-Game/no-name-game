package main

import (
	"fmt"
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

		// Recupero tutto le attività da notificare
		var rGetPlayerActivityToNotify *pb.GetPlayerActivityToNotifyResponse
		if rGetPlayerActivityToNotify, err = config.App.Server.Connection.GetPlayerActivityToNotify(helpers.NewContext(1), &pb.GetPlayerActivityToNotifyRequest{}); err != nil {
			logrus.Panic(err)
		}

		logrus.Infof("[*] Player Activity Notifications found: %d", len(rGetPlayerActivityToNotify.GetPlayerActivities()))
		for _, activity := range rGetPlayerActivityToNotify.GetPlayerActivities() {
			go handleActivityNotification(activity)
		}

		// Recupero tutti gli achievement da notificare
		var rGetPlayerAchievementToNotify *pb.GetPlayerAchievementToNotifyResponse
		if rGetPlayerAchievementToNotify, err = config.App.Server.Connection.GetPlayerAchievementToNotify(helpers.NewContext(1), &pb.GetPlayerAchievementToNotifyRequest{}); err != nil {
			logrus.Panic(err)
		}

		logrus.Infof("[*] Player Achievement Notifications found: %d", len(rGetPlayerActivityToNotify.GetPlayerActivities()))
		for _, playerAchievement := range rGetPlayerAchievementToNotify.GetPlayerAchievements() {
			go handleAchievementNotification(playerAchievement)
		}

		// Sleep for minute
		time.Sleep(SleepTimer)
	}
}

func handleAchievementNotification(playerAchievement *pb.PlayerAchievement) {
	var err error
	logrus.Infof("[*] Handle Achievement Notification: %d", playerAchievement.ID)

	defer func() {
		if err := recover(); err != nil {
			logrus.Info("[*] Achievement %d recovered", playerAchievement.ID)
		}
	}()

	// Recupero testo da notificare
	// text, _ := config.App.Localization.GetTranslation("notificaton.achievement.message", playerAchievement.GetPlayer().GetLanguage().GetSlug(),
	// 	helpers.Trans(playerAchievement.GetPlayer().GetLanguage().GetSlug(), fmt.Sprintf("achievement.%s", playerAchievement.GetAchievement().GetSlug())), // Achievement
	// 	playerAchievement.GetAchievement().GetGoldReward(),
	// 	playerAchievement.GetAchievement().GetDiamondReward(),
	// 	playerAchievement.GetAchievement().GetExperienceReward(),
	// 	)

	text := helpers.Trans(playerAchievement.GetPlayer().GetLanguage().GetSlug(), "notification.achievement.message",
		helpers.Trans(playerAchievement.GetPlayer().GetLanguage().GetSlug(), fmt.Sprintf("achievement.%s", playerAchievement.GetAchievement().GetSlug())), // Achievement
		playerAchievement.GetAchievement().GetGoldReward(),
		playerAchievement.GetAchievement().GetDiamondReward(),
		playerAchievement.GetAchievement().GetExperienceReward(),
	)

	// Invio notifica
	msg := helpers.NewMessage(playerAchievement.GetPlayer().GetChatID(), text)
	msg.ParseMode = "markdown"
	if _, err = helpers.SendMessage(msg); err != nil {
		logrus.Panic(err)
	}

	// Aggiorno lo stato levando la notifica
	if _, err = config.App.Server.Connection.SetPlayerAchievementNotified(helpers.NewContext(1), &pb.SetPlayerAchievementNotifiedRequest{
		AchievementID: playerAchievement.ID,
	}); err != nil {
		logrus.Panic(err)
	}
}

func handleActivityNotification(activity *pb.PlayerActivity) {
	var err error
	logrus.Infof("[*] Handle Activity Activity: %d", activity.ID)

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
	text := helpers.Trans(rGetPlayerByID.GetPlayer().GetLanguage().GetSlug(), fmt.Sprintf("notification.activity.%s", activity.Controller))

	msg := helpers.NewMessage(rGetPlayerByID.GetPlayer().GetChatID(), text)
	msg.ParseMode = "markdown"
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
