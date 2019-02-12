package app

import (
	"encoding/json"
	"reflect"
	"strconv"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// Helper - Check if user exist in DB, if not exist create!
func checkUser(message *tgbotapi.Message) bool {
	player = findPlayerByUsername(message.From.UserName)
	if player.ID < 1 {
		player = Player{
			Username: message.From.UserName,
			Language: getDefaultLangID("en"),
		}

		player.create()
	}

	return true
}

// Helper - Unmarshal payload state
func unmarshalPayload(payload string, funcInterface interface{}) {
	if payload != "" {
		err := json.Unmarshal([]byte(payload), &funcInterface)
		if err != nil {
			services.ErrorHandler("Error unmarshal payload", err)
		}
	}
}

// Helper - set function state in Redis
func getRedisState(player Player) string {
	var route string
	route, _ = services.Redis.Get(strconv.FormatUint(uint64(player.ID), 10)).Result()

	return route
}

// Helper - set function state in Redis
func setRedisState(player Player, function string) {
	err := services.Redis.Set(strconv.FormatUint(uint64(player.ID), 10), function, 0).Err()
	if err != nil {
		services.ErrorHandler("Error SET player state in redis", err)
	}
}

// Helper - del function state in Redis
func delRedisState(player Player) {
	err := services.Redis.Del(strconv.FormatUint(uint64(player.ID), 10)).Err()
	if err != nil {
		services.ErrorHandler("Error DEL player state in redis", err)
	}
}

// trans - late shortCut
func trans(key, locale string, args ...interface{}) (message string) {
	if len(args) <= 0 {
		message, _ = services.GetTranslation(key, locale)
		return
	}

	message, _ = services.GetTranslation(key, locale, args)
	return
}

// Helper - check if val exist in array
func inArray(val interface{}, array interface{}) (exists bool) {
	exists = false

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				exists = true
				return
			}
		}
	}

	return
}
