package helpers

import (
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/no-name/app/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StatsKeyboard - Generate a keyboard with stats
func StatsKeyboard() (keyboard tgbotapi.ReplyKeyboardMarkup) {

	var keyboardRows [][]tgbotapi.KeyboardButton

	val := reflect.ValueOf(&models.PlayerStats{}).Elem()
	for i := 3; i < val.NumField()-1; i++ {
		typeField := val.Type().Field(i)
		keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(typeField.Name))
		keyboardRows = append(keyboardRows, keyboardRow)
	}

	keyboard.Keyboard = keyboardRows
	return

}

func InStatsArray(s string) (ok bool) {
	val := reflect.ValueOf(&models.PlayerStats{}).Elem()
	for i := 3; i < val.NumField()-1; i++ {
		fieldName := val.Type().Field(i).Name
		if strings.EqualFold(fieldName, s) {
			return true
		}
	}
	return false
}
