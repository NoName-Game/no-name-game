package controllers

import (
	"fmt"
	"log"
	"strings"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
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
func (c *InventoryRecapController) Handle(player *pb.Player, update tgbotapi.Update, proxy bool) {
	var err error
	var finalRecap string

	c.Update = update

	// *******************
	// Recupero risorse inventario
	// *******************
	rGetPlayerResource, err := services.NnSDK.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}

	var recapResources string
	recapResources = fmt.Sprintf("*%s*:\n", helpers.Trans(player.Language.Slug, "resources"))
	for _, resource := range rGetPlayerResource.GetPlayerInventory() {
		log.Println("Item: ", resource.GetResource().GetName())
		recapResources += fmt.Sprintf(
			"- %v x %s (*%s*)\n",
			resource.GetQuantity(),
			resource.GetResource().GetName(),
			strings.ToUpper(resource.GetResource().GetRarity().GetSlug()),
		)
	}

	// *******************
	// Recupero item inventario
	// *******************
	rGetPlayerItems, err := services.NnSDK.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}

	var recapItems string
	recapItems = fmt.Sprintf("*%s*:\n", helpers.Trans(player.Language.Slug, "items"))
	for _, resource := range rGetPlayerItems.GetPlayerInventory() {
		log.Println("Item: ", resource.GetItem().GetName())
		recapItems += fmt.Sprintf(
			"- %v x %s (*%s*)\n",
			resource.Quantity,
			helpers.Trans(player.Language.Slug, "items."+resource.GetItem().GetName()),
			strings.ToUpper(resource.GetItem().GetRarity().GetSlug()),
		)
	}

	// Riassumo il tutto
	finalRecap = fmt.Sprintf("%s\n\n%s\n%s",
		helpers.Trans(player.Language.Slug, "inventory.recap"), // Ecco il tuo inventario
		recapResources,
		recapItems,
	)

	log.Println(finalRecap)

	msg := services.NewMessage(c.Update.Message.Chat.ID, finalRecap)
	msg.ParseMode = "markdown"

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}
