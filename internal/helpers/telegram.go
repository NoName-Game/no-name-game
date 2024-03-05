package helpers

import (
	"encoding/json"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"nn-telegram/config"
)

type InlineDataStruct struct {
	C  string // Controller
	AT string // ActionType Es. fight|movements
	A  string // Action     Es. hit
	SA string // SubAction  Es ability
	D  uint32 // Data       Es 1, 2, 3 ecc
}

func (d *InlineDataStruct) GetDataString() string {
	marshalData, err := json.Marshal(d)
	if err != nil {
		logrus.Errorf("error marshal hunting data: %s", err.Error())
	}

	return string(marshalData)
}

func (d InlineDataStruct) GetDataValue(stringData string) (data InlineDataStruct) {
	if err := json.Unmarshal([]byte(stringData), &data); err != nil {
		logrus.Errorf("error unmarshal hunting data: %s", err.Error())
	}

	return
}

// NewMessage creates a new Message.
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
	// Richiamo ratelimiter
	_ = config.App.RateLimiter.Limiter.Take()

	if message, err = config.App.Bot.API.Send(chattable); err != nil {
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

func SplitMessage(message string) (out []string) {
	buf := strings.Split(message, "\n")

	var curr string
	for _, s := range buf {
		if len(curr+" "+s) <= 2048 {
			curr += fmt.Sprintf(" %s\n", s)
		} else {
			out = append(out, curr)
			curr = ""
		}
	}

	// final result
	out = append(out, curr)
	return
}
