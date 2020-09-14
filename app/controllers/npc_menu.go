package controllers

import (
	"fmt"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Writer: reloonfire
// Starting on: 09/07/2020
// Project: nn-telegram

type NpcMenuController struct {
	BaseController
	SafePlanet bool // Flag per verificare se il player si trova su un pianeta sicuro
}

// ====================================
// Handle
// ====================================
func (c *NpcMenuController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init funzionalità
	c.Configuration.Controller = "route.menu.npc"

	// Il menù del player refresha sempre lo status del player
	var rGetPlayerByUsername *pb.GetPlayerByUsernameResponse
	if rGetPlayerByUsername, err = services.NnSDK.GetPlayerByUsername(helpers.NewContext(1), &pb.GetPlayerByUsernameRequest{
		Username: player.GetUsername(),
	}); err != nil {
		panic(err)
	}

	// Recupero dettagli utente
	c.Player = rGetPlayerByUsername.GetPlayer()

	msg := services.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.welcome"))
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       c.GetKeyboard(),
		ResizeKeyboard: true,
	}

	// Send recap message
	if _, err = services.SendMessage(msg); err != nil {
		panic(err)
	}

	// Se il player è morto non può fare altro che riposare o azioni che richiedono riposo
	if c.PlayerData.PlayerStats.GetDead() {
		restsController := new(ShipRestsController)
		restsController.Handle(c.Player, c.Update)
	}
}

func (c *NpcMenuController) GetKeyboard() [][]tgbotapi.KeyboardButton {
	// Recupero gli npc attivi in questo momento
	rGetAll, err := services.NnSDK.GetAllNPC(helpers.NewContext(1), &pb.GetAllNPCRequest{})
	if err != nil {
		panic(err)
	}

	var keyboardRow [][]tgbotapi.KeyboardButton
	for _, npc := range rGetAll.GetNPCs() {
		row := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(
				helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("route.safeplanet.%s", npc.Slug)),
			),
		)
		keyboardRow = append(keyboardRow, row)
	}

	keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.menu")),
	))

	return keyboardRow
}
