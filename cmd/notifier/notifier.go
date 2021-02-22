package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/joho/godotenv/autoload" // Autload .env
	"github.com/sirupsen/logrus"
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

		// ***************
		// Activity
		// ***************

		// Recupero tutto le attività da notificare
		var rGetPlayerActivityToNotify *pb.GetPlayerActivityToNotifyResponse
		if rGetPlayerActivityToNotify, err = config.App.Server.Connection.GetPlayerActivityToNotify(helpers.NewContext(1), &pb.GetPlayerActivityToNotifyRequest{}); err != nil {
			logrus.Panic(err)
		}

		logrus.Infof("[*] Player Activity Notifications found: %d", len(rGetPlayerActivityToNotify.GetPlayerActivities()))
		for _, activity := range rGetPlayerActivityToNotify.GetPlayerActivities() {
			go handleActivityNotification(activity)
		}

		// ***************
		// Achievements
		// ***************

		// Recupero tutti gli achievement da notificare
		var rGetPlayerAchievementToNotify *pb.GetPlayerAchievementToNotifyResponse
		if rGetPlayerAchievementToNotify, err = config.App.Server.Connection.GetPlayerAchievementToNotify(helpers.NewContext(1), &pb.GetPlayerAchievementToNotifyRequest{}); err != nil {
			logrus.Panic(err)
		}

		logrus.Infof("[*] Player Achievement Notifications found: %d", len(rGetPlayerActivityToNotify.GetPlayerActivities()))
		for _, playerAchievement := range rGetPlayerAchievementToNotify.GetPlayerAchievements() {
			go handleAchievementNotification(playerAchievement)
		}

		// ***************
		// Titan Drop
		// ***************

		// Recupero tutti i drop da notificare
		var rGetTitanDropToNotify *pb.GetTitanDropToNotifyResponse
		if rGetTitanDropToNotify, err = config.App.Server.Connection.GetTitanDropToNotify(helpers.NewContext(1), &pb.GetTitanDropToNotifyRequest{}); err != nil {
			logrus.Panic(err)
		}

		logrus.Infof("[*] Player Titan Drop Notifications found: %d", len(rGetPlayerActivityToNotify.GetPlayerActivities()))
		for _, drops := range rGetTitanDropToNotify.GetTitanDrops() {
			go handleTitanDropNotification(drops)
		}

		// Sleep for minute
		time.Sleep(SleepTimer)
	}
}

func handleTitanDropNotification(playerTitanDrop *pb.PlayerTitanDrop) {
	var err error
	logrus.Infof("[*] Handle Achievement Titan Drop: %d", playerTitanDrop.ID)

	defer func() {
		if err := recover(); err != nil {
			logrus.Info("[*] Achievement %d recovered", playerTitanDrop.ID)
		}
	}()

	// Costruisco messaggio drop base
	dropBase := helpers.Trans(playerTitanDrop.GetPlayer().GetLanguage().GetSlug(), "notification.titan_drop.base",
		playerTitanDrop.GetMoney(),
		playerTitanDrop.GetDiamond(),
		playerTitanDrop.GetExperience(),
	)

	// Costruisco messaggio drop aggiuntivo
	var dropPlus string
	if playerTitanDrop.GetWeapon() != nil {
		dropPlus = helpers.Trans(playerTitanDrop.GetPlayer().GetLanguage().GetSlug(), "notification.titan_drop.weapon",
			playerTitanDrop.GetWeapon().GetName(),
			playerTitanDrop.GetWeapon().GetRarity().GetSlug(),
		)
	} else if playerTitanDrop.GetArmor() != nil {
		var dropArmorType string
		switch playerTitanDrop.GetArmor().GetArmorCategoryID() {
		case 1:
			dropArmorType = "notification.titan_drop.armor.helmet"
		case 2:
			dropArmorType = "notification.titan_drop.armor.glove"
		case 3:
			dropArmorType = "notification.titan_drop.armor.chest"
		case 4:
			dropArmorType = "notification.titan_drop.armor.boots"
		}

		dropPlus = helpers.Trans(playerTitanDrop.GetPlayer().GetLanguage().GetSlug(), dropArmorType,
			playerTitanDrop.GetArmor().GetName(),
			playerTitanDrop.GetArmor().GetRarity().GetSlug(),
		)
	}

	// Recupero testo da notificare
	text := helpers.Trans(playerTitanDrop.GetPlayer().GetLanguage().GetSlug(), "notification.titan_drop.message",
		playerTitanDrop.GetTitan().GetName(),
		playerTitanDrop.GetDamageInflicted(),
		dropBase,
		dropPlus,
	)

	// Invio notifica
	msg := helpers.NewMessage(playerTitanDrop.GetPlayer().GetChatID(), text)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err = helpers.SendMessage(msg); err != nil {
		logrus.Panic(err)
	}

	// Aggiorno lo stato levando la notifica
	if _, err = config.App.Server.Connection.SetTitanDropNotified(helpers.NewContext(1), &pb.SetTitanDropNotifiedRequest{
		TitanDropID: playerTitanDrop.GetID(),
	}); err != nil {
		logrus.Panic(err)
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
	text := helpers.Trans(playerAchievement.GetPlayer().GetLanguage().GetSlug(), "notification.achievement.message",
		helpers.Trans(playerAchievement.GetPlayer().GetLanguage().GetSlug(), fmt.Sprintf("achievement.%s", playerAchievement.GetAchievement().GetSlug())), // Achievement
		playerAchievement.GetAchievement().GetGoldReward(),
		playerAchievement.GetAchievement().GetDiamondReward(),
		playerAchievement.GetAchievement().GetExperienceReward(),
	)

	// Invio notifica
	msg := helpers.NewMessage(playerAchievement.GetPlayer().GetChatID(), text)
	msg.ParseMode = tgbotapi.ModeHTML
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
	msg.ParseMode = tgbotapi.ModeHTML
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
