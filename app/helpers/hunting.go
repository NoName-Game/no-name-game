package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/app/provider"
	"bitbucket.org/no-name-game/no-name/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func GenerateWeaponKeyboard(player nnsdk.Player) (keyboardRows [][]tgbotapi.KeyboardButton) {
	// FIXME: remove this
	weapons, err := provider.GetPlayerWeapons(player, "true")
	if err != nil {
		services.ErrorHandler("Cant get weapon equipped", err)
	}

	for _, weapon := range weapons {
		keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(Trans("combat.attack_with", player.Language.Slug) + weapon.Name))
		keyboardRows = append(keyboardRows, keyboardRow)
	}

	return keyboardRows
}
