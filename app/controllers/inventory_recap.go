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
// Inventory Recap
// ====================================

type InventoryRecapController BaseController

// ====================================
// Handle
// ====================================
func (c *InventoryRecapController) Handle(player nnsdk.Player, update tgbotapi.Update, proxy bool) {
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
		recapResources += fmt.Sprintf(
			"- %v x %s (*%s*)\n",
			*resource.Quantity,
			resource.Resource.Name,
			strings.ToUpper(resource.Resource.Rarity.Slug),
		)
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
		recapItems += fmt.Sprintf(
			"- %v x %s (*%s*)\n",
			*resource.Quantity,
			helpers.Trans(player.Language.Slug, "items."+resource.Item.Slug),
			strings.ToUpper(resource.Item.Rarity.Slug),
		)
	}

	// Riassumo il tutto
	finalRecap = fmt.Sprintf("%s\n\n%s\n%s",
		helpers.Trans(player.Language.Slug, "inventory.recap"), // Ecco il tuo inventario
		recapResources,
		recapItems,
	)

	msg := services.NewMessage(c.Update.Message.Chat.ID, finalRecap)
	msg.ParseMode = "markdown"

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}
