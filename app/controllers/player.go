package controllers

import (
	"fmt"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Player
// ====================================
type PlayerController struct {
	Payload interface{}
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *PlayerController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	var err error

	// Verifico se è impossibile inizializzare
	if !c.InitController(
		"route.player",
		c.Payload,
		[]string{},
		player,
		update,
	) {
		return
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.CurrentState.Payload, &c.Payload)

	// Verifico se esistono condizioni per cambiare stato o uscire
	if !proxy {
		if c.BackTo(0, &MenuController{}) {
			return
		}
	}

	// Recupero armature del player
	rGetPlayerArmors, err := services.NnSDK.GetPlayerArmors(helpers.NewContext(100), &pb.GetPlayerArmorsRequest{
		PlayerID: c.Player.GetID(),
		Equipped: true,
	})
	if err != nil {
		// log.Fatalln(err)
		panic(err)
	}

	// armatura base player
	var defense, evasion, halving float32
	if len(rGetPlayerArmors.GetArmors()) > 0 {
		for _, armor := range rGetPlayerArmors.GetArmors() {
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
		c.Player.GetUsername(),
		c.PlayerStats.GetExperience(),
		c.PlayerStats.GetLevel(),
		c.PlayerStats.GetLifePoint(),
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
	// Calcolo monete del player
	responseMoney, _ := services.NnSDK.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
		PlayerID:    c.Player.GetID(),
		EconomyType: "money",
	})

	// Calcolo diamanti del player
	responseDiamond, _ := services.NnSDK.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
		PlayerID:    c.Player.GetID(),
		EconomyType: "diamond",
	})

	economy = fmt.Sprintf("💰 *%v* 💎 *%v*", responseMoney.GetValue(), responseDiamond.GetValue())

	return
}
