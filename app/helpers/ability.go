package helpers

import (
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StatsKeyboard - Generate a keyboard with stats
func StatsKeyboard() (keyboard tgbotapi.ReplyKeyboardMarkup) {
	var keyboardRows [][]tgbotapi.KeyboardButton

	val := reflect.ValueOf(&nnsdk.PlayerStats{}).Elem()
	for i := 8; i < val.NumField()-1; i++ {
		keyboardRow := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				Trans("ability." + strings.ToLower(val.Type().Field(i).Name)),
			),
		)
		keyboardRows = append(keyboardRows, keyboardRow)
	}

	keyboard.Keyboard = keyboardRows
	return

}

// InStatsStruct - Check if string exists in PlayerStats struct
func InStatsStruct(s string) (ok bool) {
	val := reflect.ValueOf(&nnsdk.PlayerStats{}).Elem()
	for i := 8; i < val.NumField()-1; i++ {
		fieldName := Trans("ability." + strings.ToLower(val.Type().Field(i).Name))
		if strings.EqualFold(fieldName, s) {
			return true
		}
	}
	return false
}
