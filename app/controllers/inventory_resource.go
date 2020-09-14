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
type InventoryResourceController struct {
	BaseController
}

// ====================================
// Handle
// ====================================
func (c *InventoryResourceController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error
	c.Update = update
	c.Player = player

	// Init Controller
	if !c.InitController(ControllerConfiguration{
		Controller: "route.inventory.resources",
		ControllerBack: ControllerBack{
			To:        &InventoryController{},
			FromStage: 0,
		},
	}) {
		return
	}

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

	// Riassumo il tutto
	var finalResource string
	finalResource = fmt.Sprintf("%s\n\n%s",
		helpers.Trans(player.Language.Slug, "inventory.recap.resource"), // Ecco il tuo inventario
		recapResources,
	)

	msg := services.NewMessage(c.Update.Message.Chat.ID, finalResource)
	msg.ParseMode = "markdown"
	if _, err = services.SendMessage(msg); err != nil {
		panic(err)
	}
}
