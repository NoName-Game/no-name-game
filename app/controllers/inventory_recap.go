package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := services.NnSDK.GetPlayerResources(ctx, &pb.GetPlayerResourcesRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}

	var playerInventoryResources []*pb.PlayerInventory
	playerInventoryResources = response.GetPlayerInventory()

	var recapResources string
	recapResources = fmt.Sprintf("*%s*:\n", helpers.Trans(player.Language.Slug, "resources"))
	for _, resource := range playerInventoryResources {
		recapResources += fmt.Sprintf(
			"- %v x %s (*%s*)\n",
			resource.Quantity,
			resource.Resource.Name,
			strings.ToUpper(resource.Resource.Rarity.Slug),
		)
	}

	// *******************
	// Recupero item inventario
	// *******************
	responseItems, err := services.NnSDK.GetPlayerItems(ctx, &pb.GetPlayerItemsRequest{
		PlayerID: c.Player.GetID(),
	})
	if err != nil {
		panic(err)
	}

	var playerInventoryItems []*pb.PlayerInventory
	playerInventoryItems = responseItems.GetPlayerInventory()

	var recapItems string
	recapItems = fmt.Sprintf("*%s*:\n", helpers.Trans(player.Language.Slug, "items"))
	for _, resource := range playerInventoryItems {
		recapItems += fmt.Sprintf(
			"- %v x %s (*%s*)\n",
			resource.Quantity,
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
