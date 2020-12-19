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
// PlayerInventoryResource
// ====================================
type PlayerInventoryResourceController struct {
	Controller
}

// ====================================
// Handle
// ====================================
func (c *PlayerInventoryResourceController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.inventory.resources",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerInventoryController{},
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

	// Ordino l'array per rarità (dal più grande al più piccolo)
	resources := helpers.SortInventoryByRarity(rGetPlayerResource.GetPlayerInventory())

	for _, resource := range resources {
		if resource.GetQuantity() > 0 {
			recapResources += fmt.Sprintf(
				"- %s %v x %s (*%s*)\n",
				helpers.GetResourceCategoryIcons(resource.GetResource().GetResourceCategoryID()),
				resource.Quantity,
				resource.Resource.Name,
				strings.ToUpper(resource.Resource.Rarity.Slug),
			)
		}
	}

	// Riassumo il tutto
	var finalResource string
	finalResource = fmt.Sprintf("%s\n\n%s",
		helpers.Trans(player.Language.Slug, "inventory.recap.resource"), // Ecco il tuo inventario
		recapResources,
	)

	msg := helpers.NewMessage(c.Update.Message.Chat.ID, finalResource)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.more")),
		),
	)

	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *PlayerInventoryResourceController) Validator() bool {
	return false
}

func (c *PlayerInventoryResourceController) Stage() {
	//
}
