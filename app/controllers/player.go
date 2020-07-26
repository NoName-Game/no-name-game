package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Player
// ====================================
type PlayerController struct {
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *PlayerController) Handle(player nnsdk.Player, update tgbotapi.Update, proxy bool) {
	var err error
	c.Player = player
	c.Update = update
	c.Controller = "route.player"

	// Se tutto ok imposto e setto il nuovo stato su redis
	_ = helpers.SetRedisState(c.Player, c.Controller)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &MenuController{}) {
			return
		}
	}

	// Recupero armature del player
	var playerProvider providers.PlayerProvider
	var armors nnsdk.Armors
	armors, err = playerProvider.GetPlayerArmors(c.Player, "true")
	if err != nil {
		panic(err)
	}

	// armatura base player
	var defense, evasion, halving float32
	if len(armors) > 0 {
		for _, armor := range armors {
			defense += armor.Defense
			evasion += armor.Evasion
			halving += armor.Halving
		}
	}

	// Calcolo lato economico del player
	var economy string
	economy, err = c.GetPlayerEconomy()
	if err != nil {
		panic(err)
	}

	armorsEquipment := fmt.Sprintf(""+
		"👨🏼‍🚀 %s \n"+
		"🏵 *%v* 🎖 *%v* \n"+
		"♥️ *%v*/100 HP\n"+
		"🛡 Def: *%v* | Evs: *%v* | Hlv: *%v*\n"+
		"%s",
		c.Player.Username,
		c.Player.Stats.Experience,
		c.Player.Stats.Level,
		*c.Player.Stats.LifePoint,
		defense, evasion, halving,
		economy,
	)

	// msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "player.intro"))
	msg := services.NewMessage(c.Update.Message.Chat.ID, armorsEquipment)
	msg.ParseMode = "markdown"
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.ability")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory.equip")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
		),
	)

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}

func (c *PlayerController) Validator() {
	//
}

func (c *PlayerController) Stage() {
	//
}

// GetPlayerTask
// Metodo didicato alla reppresenteazione del risorse econimiche del player
func (c *PlayerController) GetPlayerEconomy() (economy string, err error) {
	var playerProvider providers.PlayerProvider

	// Calcolo monete del player
	var money nnsdk.MoneyResponse
	money, _ = playerProvider.GetPlayerEconomy(c.Player.ID, "money")

	var diamond nnsdk.MoneyResponse
	diamond, _ = playerProvider.GetPlayerEconomy(c.Player.ID, "diamond")

	economy = fmt.Sprintf("💰 *%v* 💎 *%v*", money.Value, diamond.Value)

	return
}
