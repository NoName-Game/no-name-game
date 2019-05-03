package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func GenerateWeaponKeyboard(player models.Player) (keyboardRows [][]tgbotapi.KeyboardButton) {

	for _, weapon := range player.Weapons {
		keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(Trans("combat.attack_with", player.Language.Slug) + weapon.Name))
		keyboardRows = append(keyboardRows, keyboardRow)
	}

	return keyboardRows
}
