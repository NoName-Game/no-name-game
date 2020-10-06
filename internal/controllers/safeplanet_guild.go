package controllers

import (
	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetProtectorsController
// ====================================
type SafePlanetProtectorsController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetProtectorsController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.coalition.protectors",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCoalitionController{},
				FromStage: 0,
			},
		},
	}) {
		return
	}

	// Verifico/Recupero Gilda player
	var err error
	var rGetPlayerGuild *pb.GetPlayerGuildResponse
	if rGetPlayerGuild, err = config.App.Server.Connection.GetPlayerGuild(helpers.NewContext(1), &pb.GetPlayerGuildRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	var keyboardRows [][]tgbotapi.KeyboardButton
	if !rGetPlayerGuild.GetInGuild() {
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.create")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.join")),
		))
	} else {
		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.safeplanet.coalition.protectors.leave")),
		))
	}

	// Aggiungo anche abbandona
	keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(
			helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
		),
	))

	// Message
	var protectorsMessage string
	protectorsMessage = helpers.Trans(player.Language.Slug, "safeplanet.coalition.protectors.info")
	if rGetPlayerGuild.GetInGuild() {
		protectorsMessage += helpers.Trans(player.Language.Slug, "safeplanet.coalition.protectors.player_protectors_details", rGetPlayerGuild.GetGuild().GetName())
	}

	msg := helpers.NewMessage(c.Update.Message.Chat.ID, protectorsMessage)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       keyboardRows,
		ResizeKeyboard: true,
	}

	if _, err := helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *SafePlanetProtectorsController) Validator() bool {
	return false
}

func (c *SafePlanetProtectorsController) Stage() {
	//
}
