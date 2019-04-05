package helpers

import (
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/no-name/services"

	"bitbucket.org/no-name-game/no-name/app/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StatsKeyboard - Generate a keyboard with stats
func StatsKeyboard(slug string) (keyboard tgbotapi.ReplyKeyboardMarkup) {

	var keyboardRows [][]tgbotapi.KeyboardButton

	val := reflect.ValueOf(&models.PlayerStats{}).Elem()
	for i := 3; i < val.NumField()-1; i++ {
		text, _ := services.GetTranslation("ability."+strings.ToLower(val.Type().Field(i).Name), slug, nil)
		keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(text))
		keyboardRows = append(keyboardRows, keyboardRow)
	}

	keyboard.Keyboard = keyboardRows
	return

}

// InStatsStruct - Check if string exists in PlayerStats struct
func InStatsStruct(s string, languageSlug string) (ok bool) {
	val := reflect.ValueOf(&models.PlayerStats{}).Elem()
	for i := 3; i < val.NumField()-1; i++ {
		fieldName, _ := services.GetTranslation("ability."+strings.ToLower(val.Type().Field(i).Name), languageSlug, nil)
		if strings.EqualFold(fieldName, s) {
			return true
		}
	}
	return false
}
