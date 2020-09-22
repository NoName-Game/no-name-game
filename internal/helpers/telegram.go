package helpers

import (
	"bitbucket.org/no-name-game/nn-telegram/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// NewMessage creates a new Message.
//
// chatID is where to send it, text is the message text.
func NewMessage(chatID int64, text string) tgbotapi.MessageConfig {
	return tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           chatID,
			ReplyToMessageID: 0,
		},
		Text:                  text,
		DisableWebPagePreview: false,
	}
}

// DeleteMessage - Cancella un messaggio
func DeleteMessage(chatID int64, messageID int) (err error) {
	_, err = config.App.Bot.API.DeleteMessage(tgbotapi.DeleteMessageConfig{
		ChatID:    chatID,
		MessageID: messageID,
	})

	return
}

// EditMessage - Modifica messaggio
func NewEditMessage(chatID int64, messageID int, text string) tgbotapi.EditMessageTextConfig {
	return tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    chatID,
			MessageID: messageID,
		},
		Text: text,
	}
}

// SendMessage - Invia messaggio
func SendMessage(chattable tgbotapi.Chattable) (message tgbotapi.Message, err error) {
	message, err = config.App.Bot.API.Send(chattable)
	if err != nil {
		return message, err
	}

	return
}

func NewAnswer(callbackQueryID string, text string, alert bool) tgbotapi.CallbackConfig {
	return tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQueryID,
		Text:            text,
		ShowAlert:       alert,
	}
}

func AnswerCallbackQuery(tconfig tgbotapi.CallbackConfig) (err error) {
	_, err = config.App.Bot.API.AnswerCallbackQuery(tconfig)
	if err != nil {
		return err
	}

	return
}
