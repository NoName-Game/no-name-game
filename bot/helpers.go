package bot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
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
		Text: text,
		DisableWebPagePreview: false,
	}
}

// NewDeleteMessage creates a request to delete a message.
func NewDeleteMessage(chatID int64, messageID int) tgbotapi.DeleteMessageConfig {
	return tgbotapi.DeleteMessageConfig{
		ChatID:    chatID,
		MessageID: messageID,
	}
}

// NewUpdate gets updates since the last Offset.
//
// offset is the last Update ID to include.
// You likely want to set this to the last Update ID plus 1.
func NewUpdate(offset int) tgbotapi.UpdateConfig {
	return tgbotapi.UpdateConfig{
		Offset:  offset,
		Limit:   0,
		Timeout: 0,
	}
}
