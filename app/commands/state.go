package commands

// CheckFinishTime - Check the ending and handle the functions.
func CheckFinishTime() {
	// states, _ := providers.GetPlayerStateToNotify()
	//
	// for _, state := range states {
	// 	player, _ := providers.GetPlayerByID(state.PlayerID)
	// 	text, _ := services.GetTranslation("cron."+state.Function+"_alert", player.Language.Slug, nil)
	//
	// 	// Send notification
	// 	msg := services.NewMessage(player.ChatID, text)
	// 	continueButton, _ := services.GetTranslation(state.Function, player.Language.Slug, nil)
	// 	// I need this continue button to recall the function.
	// 	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(continueButton)))
	// 	services.SendMessage(msg)
	//
	// 	// Update status
	// 	// Stupid poninter stupid json pff
	// 	f := new(bool)
	// 	*f = false
	//
	// 	state.ToNotify = f
	// 	state, _ = providers.UpdatePlayerState(state)
	// }
}
