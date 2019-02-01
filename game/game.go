package game

import (
	"log"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	"gitlab.com/Valkyrie00/no-name/bot"
	"gitlab.com/Valkyrie00/no-name/config"
)

var (
	player Player
)

func init() {
	bootstrap()
}

// Run - The Game!
func Run() {
	updates := bot.GetUpdates()
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if update.Message.From.UserName == "" {
			msg := bot.NewMessage(update.Message.Chat.ID, "Non puoi continuare senza useraname")
			bot.SendMessage(msg)
			continue
		}

		// Block to test DB
		// First or create
		player.findByUsername(update.Message.From.UserName)
		if player.ID < 1 {
			player = Player{Username: update.Message.From.UserName}
			player.create()
		}

		log.Println(player)
		// ***************************

		testMultistate(update.Message)

		msg := bot.NewMessage(update.Message.Chat.ID, update.Message.Text)
		bot.SendMessage(msg)
	}
}

// PlayerState -
type PlayerState struct {
	gorm.Model
	PlayerID int
	Function string
	Payload  string
}

// Create player
func (s *PlayerState) create() *PlayerState {
	config.Database.Create(s)

	return s
}

// Multistate
func testMultistate(message *tgbotapi.Message) {
	if message.Text == "sintesi" {
		sintesi(message)
	}
}

func sintesi(message *tgbotapi.Message) {

	var state PlayerState
	config.Database.Where("player_id = ?", message.From.ID).First(&state)

	if state.ID < 1 {
		state = PlayerState{PlayerID: message.From.ID}
		state.create()
	}

	// Setto che l'utente è entrato in sintesi come stato
	state.Function = "sintesi"
	config.Database.Save(&state)

	// Validation
}
