package helpers

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
// func ListItemsFilteredBy(items nnsdk.Crafts, rarityID uint) tgbotapi.MessageConfig {
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
// 		if item.Item.RarityID <= rarityID {
// 			row := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(Trans("crafting.craft") + item.Item.Name))
// 			keyboardRow = append(keyboardRow, row)
// 		}
// 	}
// 	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
// 		Keyboard:        keyboardRow,
// 		ResizeKeyboard:  false,
// 	}
//
// 	return msg
// }
//
// func ListRecipe(needed map[uint]int) string {
// 	var result string
//
// 	for itemID, value := range needed {
// 		var item nnsdk.Item
// 		item, err := providers.GetItemByID(itemID)
//
// 		if err != nil {
// 			services.ErrorHandler("Can't retrieve Item", err)
// 		}
//
// 		result += item.Name + " x" + strconv.Itoa(value) + "\n"
// 	}
//
// 	return result
// }

func CheckForItems(inventory, needed map[uint]int) (hasItems bool) {

	for itemId, value := range needed {
		if inventory[itemId] < value {
			return false
		}
	}

	return true
}
