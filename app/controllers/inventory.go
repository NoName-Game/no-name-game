package controllers

import (
	"fmt"
	"strings"

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
	var finalRecap string
	var playerProvider providers.PlayerProvider

	c.Update = update

	// *******************
	// Recupero risorse inventario
	// *******************

	var playerInventoryResources nnsdk.PlayerInventories
	playerInventoryResources, err = playerProvider.GetPlayerResources(player.ID)
	if err != nil {
		panic(err)
	}

	var recapResources string
	recapResources = fmt.Sprintf("*%s*:\n", helpers.Trans(player.Language.Slug, "resources"))
	for _, resource := range playerInventoryResources {
		recapResources += fmt.Sprintf("- %s x %v (*%s*)\n", resource.Resource.Name, *resource.Quantity, strings.ToUpper(resource.Resource.Rarity.Slug))
	}

	// *******************
	// Recupero item inventario
	// *******************
	var playerInventoryItems nnsdk.PlayerInventories
	playerInventoryItems, err = playerProvider.GetPlayerItems(player.ID)
	if err != nil {
		panic(err)
	}

	var recapItems string
	recapItems = fmt.Sprintf("*%s*:\n", helpers.Trans(player.Language.Slug, "items"))
	for _, resource := range playerInventoryItems {
		recapItems += fmt.Sprintf("- %s x %v (*%s*)\n", helpers.Trans(player.Language.Slug, "items."+resource.Item.Slug), *resource.Quantity, strings.ToUpper(resource.Item.Rarity.Slug))
	}

	// *******************
	// Weapons
	// *******************
	var playerWeapons nnsdk.Weapons
	playerWeapons, err = playerProvider.GetPlayerWeapons(player, "false")
	if err != nil {
		panic(err)
	}

	var recapWeapons string
	recapWeapons = fmt.Sprintf("*%s:*\n", helpers.Trans(player.Language.Slug, "weapons"))
	for _, weapon := range playerWeapons {
		recapWeapons += fmt.Sprintf("- %s (*%s*)\n", weapon.Name, strings.ToUpper(weapon.Rarity.Slug))
	}

	// *******************
	// Summary Armors
	// *******************
	var playerArmors nnsdk.Armors
	playerArmors, err = playerProvider.GetPlayerArmors(player, "false")
	if err != nil {
		panic(err)
	}

	var recapArmors string
	recapArmors = fmt.Sprintf("*%s:*\n", helpers.Trans(player.Language.Slug, "armors"))
	for _, armor := range playerArmors {
		recapArmors += fmt.Sprintf("- %s (*%s*)\n", armor.Name, strings.ToUpper(armor.Rarity.Slug))
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
