package controllers

import (
	"fmt"
	"strings"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// InventoryResource
// ====================================

type InventoryResourceController BaseController

// ====================================
// Handle
// ====================================
func (c *InventoryResourceController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error
	var finalResource string
	c.Update = update

	// *******************
	// Recupero risorse inventario
	// *******************
	var rGetPlayerResource *pb.GetPlayerResourcesResponse
	rGetPlayerResource, err = services.NnSDK.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
		PlayerID: player.GetID(),
	})
	if err != nil {
		panic(err)
	}

	var recapResources string
	for _, resource := range rGetPlayerResource.GetPlayerInventory() {
		recapResources += fmt.Sprintf(
			"- %v x %s (*%s*)\n",
			resource.Quantity,
			resource.Resource.Name,
			strings.ToUpper(resource.Resource.Rarity.Slug),
		)
	}

	// // *******************
	// // Recupero item inventario
	// // *******************
	// rGetPlayerItems, err := services.NnSDK.GetPlayerItems(helpers.NewContext(1), &pb.GetPlayerItemsRequest{
	// 	PlayerID: player.GetID(),
	// })
	// if err != nil {
	// 	panic(err)
	// }
	//
	// var recapItems string
	// recapItems = fmt.Sprintf("*%s*:\n", helpers.Trans(player.Language.Slug, "items"))
	// for _, resource := range rGetPlayerItems.GetPlayerInventory() {
	// 	recapItems += fmt.Sprintf(
	// 		"- %v x %s (*%s*)\n",
	// 		resource.Quantity,
	// 		helpers.Trans(player.Language.Slug, "items."+resource.Item.Slug),
	// 		helpers.Trans(player.Language.Slug, "items."+resource.Item.ItemCategory.Slug),
	// 	)
	// }

	// Riassumo il tutto
	finalResource = fmt.Sprintf("%s\n\n%s",
		helpers.Trans(player.Language.Slug, "inventory.recap.resource"), // Ecco il tuo inventario
		recapResources,
	)

	msg := services.NewMessage(c.Update.Message.Chat.ID, finalResource)
	msg.ParseMode = "markdown"

	_, err = services.SendMessage(msg)
	if err != nil {
		panic(err)
	}
}
