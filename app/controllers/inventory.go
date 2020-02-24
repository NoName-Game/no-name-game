package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Inventory
// ====================================
type InventoryController BaseController

// ====================================
// Handle
// ====================================
func (c *InventoryController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	var err error
	c.Update = update

	msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(player.Language.Slug, "inventory.intro"))
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory.recap")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory.items")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory.equip")),
			// tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.inventory.destroy")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.clears")),
		),
	)

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}

	return
}

// ====================================
// Inventory Recap
// ====================================

type InventoryRecapController BaseController

// ====================================
// Handle
// ====================================
func (c *InventoryRecapController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	var err error
	c.Update = update

	var finalRecap string

	// *******************
	// Recupero risorse inventario
	// *******************
	var playerInventoryResources nnsdk.PlayerInventories
	playerInventoryResources, err = providers.GetPlayerResources(player.ID)
	if err != nil {
		panic(err)
	}

	var recapResources string
	recapResources = fmt.Sprintf("*%s*:\n", helpers.Trans(player.Language.Slug, "resources"))
	for _, resource := range playerInventoryResources {
		recapResources += fmt.Sprintf("- %s x %v \n", resource.Resource.Name, *resource.Quantity)
	}

	// *******************
	// Recupero item inventario
	// *******************
	var playerInventoryItems nnsdk.PlayerInventories
	playerInventoryItems, err = providers.GetPlayerItems(player.ID)
	if err != nil {
		panic(err)
	}

	var recapItems string
	recapItems = fmt.Sprintf("*%s*:\n", helpers.Trans(player.Language.Slug, "items"))
	for _, resource := range playerInventoryItems {
		recapItems += fmt.Sprintf("- %s x %v \n", helpers.Trans(player.Language.Slug, "items."+resource.Item.Slug), *resource.Quantity)
	}

	// *******************
	// Weapons
	// *******************
	var playerWeapons nnsdk.Weapons
	playerWeapons, err = providers.GetPlayerWeapons(player, "false")
	if err != nil {
		panic(err)
	}

	var recapWeapons string
	recapWeapons = fmt.Sprintf("*%s:*\n", helpers.Trans(player.Language.Slug, "weapons"))
	for _, weapon := range playerWeapons {
		recapWeapons += fmt.Sprintf("- %s \n", weapon.Name)
	}

	// *******************
	// Summary Armors
	// *******************
	var playerArmors nnsdk.Armors
	playerArmors, err = providers.GetPlayerArmors(player, "false")
	if err != nil {
		panic(err)
	}

	var recapArmors string
	recapArmors = fmt.Sprintf("*%s:*\n", helpers.Trans(player.Language.Slug, "armors"))
	for _, armor := range playerArmors {
		recapArmors += fmt.Sprintf("- %s \n", armor.Name)
	}

	// Riassumo il tutto
	finalRecap = fmt.Sprintf("%s \n %s \n %s \n %s \n %s",
		helpers.Trans(player.Language.Slug, "inventory.recap"), // Ecco il tuo inventario
		recapResources,
		recapItems,
		recapWeapons,
		recapArmors,
	)

	msg := services.NewMessage(c.Update.Message.Chat.ID, finalRecap)
	msg.ParseMode = "markdown"

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}

	return
}
