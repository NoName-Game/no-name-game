package helpers

import (
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/app/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// StatsKeyboard - Generate a keyboard with stats
func StatsKeyboard() (keyboard tgbotapi.ReplyKeyboardMarkup, slug string) {

	var keyboardRows [][]tgbotapi.KeyboardButton

	val := reflect.ValueOf(&models.PlayerStats{}).Elem()
	for i := 3; i < val.NumField()-1; i++ {
		keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans("ability."+strings.ToLower(val.Type().Field(i).Name), slug)))
		keyboardRows = append(keyboardRows, keyboardRow)
	}

	keyboard.Keyboard = keyboardRows
	return

}

func InStatsArray(s string, slug string) (ok bool) {
	val := reflect.ValueOf(&models.PlayerStats{}).Elem()
	for i := 3; i < val.NumField()-1; i++ {
		fieldName := helpers.Trans("ability."+strings.ToLower(val.Type().Field(i).Name), slug)
		if strings.EqualFold(fieldName, s) {
			return true
		}
	}
	return false
}
