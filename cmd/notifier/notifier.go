package main

import (
	"encoding/json"
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
		// Notifications
		// ***************

		// Recupero tutte le notifiche
		var rGetNotifications *pb.GetNotificationsResponse
		if rGetNotifications, err = config.App.Server.Connection.GetNotifications(helpers.NewContext(1), &pb.GetNotificationsRequest{}); err != nil {
			logrus.Panic(err)
		}

		logrus.Infof("[*] Player Notifications found: %d", len(rGetPlayerActivityToNotify.GetPlayerActivities()))
		for _, notification := range rGetNotifications.GetNotifications() {
			go handleNotification(notification)
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

func handleNotification(notification *pb.Notification) {
	var err error
	var message string

	defer func() {
		if err := recover(); err != nil {
			logrus.Infof("[*] Notification %v recovered", notification.ID)

			// Setto messaggio come notificato
			if _, err = config.App.Server.Connection.SetNotificationNotified(helpers.NewContext(1), &pb.SetNotificationNotifiedRequest{
				NotificationID: notification.GetID(),
			}); err != nil {
				logrus.Panic(err)
			}
		}
	}()

	switch notification.GetNotificationCategory().GetSlug() {
	case "level":
		type LevelNotificationPayload struct {
			LevelID uint32
		}

		var payload LevelNotificationPayload
		_ = json.Unmarshal([]byte(notification.GetPayload()), &payload)

		message = helpers.Trans(notification.GetPlayer().GetLanguage().GetSlug(), "notification.level.message", payload.LevelID)
	case "rank":
		type RankNotificationPayload struct {
			RankID   uint32
			NameCode string
		}

		var payload RankNotificationPayload
		_ = json.Unmarshal([]byte(notification.GetPayload()), &payload)

		message = helpers.Trans(notification.GetPlayer().GetLanguage().GetSlug(), "notification.rank.message",
			helpers.Trans(notification.GetPlayer().GetLanguage().GetSlug(), fmt.Sprintf("rank.%s", payload.NameCode)),
		)
	case "achievements":
		type AchievementNotificationPayload struct {
			AchievementID uint32
		}

		var payload AchievementNotificationPayload
		_ = json.Unmarshal([]byte(notification.GetPayload()), &payload)

		// Recupero dettagli achievement
		var rGetAchievementByID *pb.GetAchievementByIDResponse
		if rGetAchievementByID, err = config.App.Server.Connection.GetAchievementByID(helpers.NewContext(1), &pb.GetAchievementByIDRequest{
			AchievementID: payload.AchievementID,
		}); err != nil {
			logrus.Panic(err)
		}

		// Recupero testo da notificare
		message = helpers.Trans(notification.GetPlayer().GetLanguage().GetSlug(), "notification.achievement.message",
			helpers.Trans(notification.GetPlayer().GetLanguage().GetSlug(), fmt.Sprintf("achievement.%s", rGetAchievementByID.GetAchievement().GetSlug())), // Achievement
			rGetAchievementByID.GetAchievement().GetGoldReward(),
			rGetAchievementByID.GetAchievement().GetDiamondReward(),
			rGetAchievementByID.GetAchievement().GetExperienceReward(),
		)
	}

	// Invio notifica
	msg := helpers.NewMessage(notification.GetPlayer().GetChatID(), message)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err = helpers.SendMessage(msg); err != nil {
		logrus.Panic(err)
	}

	// Setto messaggio come notificato
	if _, err = config.App.Server.Connection.SetNotificationNotified(helpers.NewContext(1), &pb.SetNotificationNotifiedRequest{
		NotificationID: notification.GetID(),
	}); err != nil {
		logrus.Panic(err)
	}
}

func handleTitanDropNotification(playerTitanDrop *pb.PlayerTitanDrop) {
	var err error
	logrus.Infof("[*] Handle Achievement Titan Drop: %d", playerTitanDrop.ID)

	defer func() {
		if err := recover(); err != nil {
			logrus.Info("[*] Achievement %d recovered", playerTitanDrop.ID)

			// Aggiorno lo stato levando la notifica
			if _, err = config.App.Server.Connection.SetTitanDropNotified(helpers.NewContext(1), &pb.SetTitanDropNotifiedRequest{
				TitanDropID: playerTitanDrop.GetID(),
			}); err != nil {
				logrus.Panic(err)
			}
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
