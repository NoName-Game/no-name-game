package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerPartyController
// ====================================
type PlayerPartyController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerPartyController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.party",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerController{},
				FromStage: 0,
			},
		},
	}) {
		return
	}

	// Recupero party player
	var rGetPartyDetails *pb.GetPartyDetailsResponse
	rGetPartyDetails, _ = config.App.Server.Connection.GetPartyDetails(helpers.NewContext(1), &pb.GetPartyDetailsRequest{
		PlayerID: c.Player.ID,
	})

	// Se il player si trova in un party recupero i dettagli
	if rGetPartyDetails.GetInParty() {
		// Ciclio utenti nel party
		var playerRecap string
		for _, player := range rGetPartyDetails.GetPlayers() {
			// Recupero posizione player
			var currentPosition *pb.Planet
			if currentPosition, err = helpers.GetPlayerPosition(player.ID); err != nil {
				c.Logger.Panic(err)
			}

			playerRecap += fmt.Sprintf("- <b>%s</b> [🌏 %s]\n", player.GetUsername(), currentPosition.GetName())
		}

		// Costruisco tastiera gestione party
		var partysKeyboard [][]tgbotapi.KeyboardButton
		partysKeyboard = append(partysKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.party.leave")),
		))

		// Aggiungo tasti gestione party se owner
		if rGetPartyDetails.GetOwner().GetID() == c.Player.ID {
			partysKeyboard = append(partysKeyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.party.add_player")),
				tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.party.remove_player")),
			))
		}

		// Aggiungo torna indietro
		partysKeyboard = append(partysKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "player.party.show",
			rGetPartyDetails.GetOwner().GetUsername(),
			rGetPartyDetails.GetNPlayers(),
			playerRecap,
		))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       partysKeyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}
		return
	}

	// Il Player non è in un party
	msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "player.party.non_in_party"))
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.player.party.create")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
		),
	)

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *PlayerPartyController) Validator() bool {
	return false
}

func (c *PlayerPartyController) Stage() {
	//
}
