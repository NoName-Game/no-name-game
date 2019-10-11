package helpers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func GenerateWeaponKeyboard() (keyboardRows [][]tgbotapi.KeyboardButton) {
	// FIXME: remove this
	weapons, err := providers.GetPlayerWeapons(Player, "true")
	if err != nil {
		services.ErrorHandler("Cant get weapon equipped", err)
	}

	for _, weapon := range weapons {
		keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(Trans("combat.attack_with", weapon.Name)))
		keyboardRows = append(keyboardRows, keyboardRow)
	}

	return keyboardRows
}

func GetHuntingMap(IDMap uint, player nnsdk.Player) (huntingMap nnsdk.Map, isNew bool) {
	if IDMap > 0 {
		huntingMap = GetHuntingRedisState(IDMap, player)
		return huntingMap, false
	}

	// Se IDMap non viene passato genero nuova mappa
	huntingMap, _ = providers.CreateMap(player.ID)
	SetHuntingRedisState(huntingMap.ID, player, huntingMap)
	return huntingMap, true
}

func UpdateHuntingMap(huntingMap nnsdk.Map, player nnsdk.Player) {
	SetHuntingRedisState(huntingMap.ID, player, huntingMap)
	return
}
