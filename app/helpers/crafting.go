package helpers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Writer: reloonfire
// Starting on: 19/01/2020
// Project: no-name-game

func ListItems(items nnsdk.Crafts) tgbotapi.MessageConfig {

	var result string

	for _, item := range items {
		result += item.Item.Name + "\n"
	}

	msg := tgbotapi.NewMessage(Player.ChatID, result)

	var keyboardRow [][]tgbotapi.KeyboardButton
	for _, item := range items {
		row := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(Trans("crafting.craft") + item.Item.Name))
		keyboardRow = append(keyboardRow, row)
	}
	msg.ReplyMarkup = keyboardRow

	return msg
}

func ListRecipe(needed map[uint]int) string {
	var result string

	for itemID, value := range needed {
		var item nnsdk.Item
		providers.GetItem
	}
}
