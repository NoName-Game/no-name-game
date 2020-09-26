package controllers

import (
	"fmt"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// InventoryResource
// ====================================
type InventoryResourceController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *InventoryResourceController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.inventory.resources",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &InventoryController{},
				FromStage: 0,
			},
		},
	}) {
		return
	}

	// *******************
	// Recupero risorse inventario
	// *******************
	var rGetPlayerResource *pb.GetPlayerResourcesResponse
	if rGetPlayerResource, err = config.App.Server.Connection.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
		PlayerID: player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
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

	msg := helpers.NewMessage(c.Update.Message.Chat.ID, finalResource)
	msg.ParseMode = "markdown"
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *InventoryResourceController) Validator() bool {
	return false
}

func (c *InventoryResourceController) Stage() {
	//
}
