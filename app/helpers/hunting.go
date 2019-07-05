package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/provider"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func GenerateWeaponKeyboard() (keyboardRows [][]tgbotapi.KeyboardButton) {
	// FIXME: remove this
	weapons, err := provider.GetPlayerWeapons(Player, "true")
	if err != nil {
		services.ErrorHandler("Cant get weapon equipped", err)
	}

	for _, weapon := range weapons {
		keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(Trans("combat.attack_with", weapon.Name)))
		keyboardRows = append(keyboardRows, keyboardRow)
	}

	return keyboardRows
}
