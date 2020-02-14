package helpers

import (
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
)

// Writer: reloonfire
// Starting on: 19/01/2020
// Project: no-name-game

// func ListItems(items nnsdk.Crafts) tgbotapi.MessageConfig {
//
// 	var result string
//
// 	for _, item := range items {
// 		result += item.Item.Name + "\n"
// 	}
//
// 	msg := services.NewMessage(Player.ChatID, "Ecco cosa puoi craftare:")
//
// 	var keyboardRow [][]tgbotapi.KeyboardButton
// 	for _, item := range items {
// 		row := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(Trans("crafting.craft") + item.Item.Name))
// 		keyboardRow = append(keyboardRow, row)
// 	}
// 	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
// 		Keyboard:        keyboardRow,
// 		ResizeKeyboard:  false,
// 	}
//
// 	return msg
// }
//

// ListRecipe - Metodo che aiuta a recuperare la lista di risore necessarie
// al crafting di un determianto item
func ListRecipe(needed map[uint]int) (result string, err error) {
	for resourceID, value := range needed {
		var resource nnsdk.Resource
		resource, err := providers.GetResourceByID(resourceID)
		if err != nil {
			return result, err
		}

		result += resource.Name + " x" + strconv.Itoa(value) + "\n"
	}

	return result, err
}

func CheckForItems(inventory, needed map[uint]int) (hasItems bool) {

	for itemId, value := range needed {
		if inventory[itemId] < value {
			return false
		}
	}

	return true
}
