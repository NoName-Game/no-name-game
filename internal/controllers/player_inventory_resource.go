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

func (c *PlayerInventoryResourceController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
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
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *PlayerInventoryResourceController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// *******************
	// Recupero risorse inventario
	// *******************
	var resourcesRecap string
	switch c.Update.Message.Text {
	case helpers.Trans(c.Player.GetLanguage().GetSlug(), "inventory.recap.underground"):
		resourcesRecap = c.GetResourcesUnderground()
	case helpers.Trans(c.Player.GetLanguage().GetSlug(), "inventory.recap.surface"):
		resourcesRecap = c.GetResourcesSurface()
	case helpers.Trans(c.Player.GetLanguage().GetSlug(), "inventory.recap.atmosphere"):
		resourcesRecap = c.GetResourcesAtmosphere()
	default:
		resourcesRecap = c.GetResourcesRecap()
	}

	// Riassumo il tutto
	var finalResource string
	finalResource = fmt.Sprintf("%s\n\n%s",
		helpers.Trans(player.Language.Slug, "inventory.recap.resource"), // Ecco il tuo inventario
		resourcesRecap,
	)

	for _, text := range helpers.SplitMessage(finalResource) {
		if text != "" {
			msg := helpers.NewMessage(c.ChatID, text)
			msg.ParseMode = tgbotapi.ModeHTML
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "inventory.recap.underground")),
					tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "inventory.recap.surface")),
					tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "inventory.recap.atmosphere")),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(player.Language.Slug, "route.breaker.menu")),
				),
			)

			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}
		}
	}
}

func (c *PlayerInventoryResourceController) Validator() bool {
	return false
}

func (c *PlayerInventoryResourceController) Stage() {
	//
}

func (c *PlayerInventoryResourceController) GetResourcesRecap() string {
	var err error
	var rGetPlayerResource *pb.GetPlayerResourcesResponse
	if rGetPlayerResource, err = config.App.Server.Connection.GetPlayerResources(helpers.NewContext(1), &pb.GetPlayerResourcesRequest{
		PlayerID: c.Player.GetID(),
	}); err != nil {
		c.Logger.Panic(err)
	}

	var undergroundCounter, surfaceCounter, atmosphereCounter int32
	var vcCounter, cCounter, uCounter, rCounter, urCounter, lCoutner int32
	for _, inventory := range rGetPlayerResource.GetPlayerInventory() {
		if inventory.Quantity > 0 {
			switch inventory.GetResource().GetResourceCategoryID() {
			case 1:
				undergroundCounter++
			case 2:
				surfaceCounter++
			case 3:
				atmosphereCounter++
			}

			switch inventory.GetResource().GetRarityID() {
			case 1:
				vcCounter += inventory.GetQuantity()
			case 2:
				cCounter += inventory.GetQuantity()
			case 3:
				uCounter += inventory.GetQuantity()
			case 4:
				rCounter += inventory.GetQuantity()
			case 5:
				urCounter += inventory.GetQuantity()
			case 6:
				lCoutner += inventory.GetQuantity()
			}
		}
	}

	return helpers.Trans(c.Player.GetLanguage().GetSlug(), "inventory.recap.all",
		undergroundCounter, surfaceCounter, atmosphereCounter,
		vcCounter, cCounter, uCounter, rCounter, urCounter, lCoutner,
	)
}

func (c *PlayerInventoryResourceController) GetResourcesUnderground() string {
	var err error

	var rGetPlayerResource *pb.GetPlayerResourcesByCategoryIDResponse
	if rGetPlayerResource, err = config.App.Server.Connection.GetPlayerResourcesByCategoryID(helpers.NewContext(1), &pb.GetPlayerResourcesByCategoryIDRequest{
		PlayerID:           c.Player.GetID(),
		ResourceCategoryID: 1,
	}); err != nil {
		c.Logger.Panic(err)
	}

	return c.GetResourceList(rGetPlayerResource.GetPlayerInventory())
}

func (c *PlayerInventoryResourceController) GetResourcesSurface() string {
	var err error

	var rGetPlayerResource *pb.GetPlayerResourcesByCategoryIDResponse
	if rGetPlayerResource, err = config.App.Server.Connection.GetPlayerResourcesByCategoryID(helpers.NewContext(1), &pb.GetPlayerResourcesByCategoryIDRequest{
		PlayerID:           c.Player.GetID(),
		ResourceCategoryID: 2,
	}); err != nil {
		c.Logger.Panic(err)
	}

	return c.GetResourceList(rGetPlayerResource.GetPlayerInventory())
}

func (c *PlayerInventoryResourceController) GetResourcesAtmosphere() string {
	var err error

	var rGetPlayerResource *pb.GetPlayerResourcesByCategoryIDResponse
	if rGetPlayerResource, err = config.App.Server.Connection.GetPlayerResourcesByCategoryID(helpers.NewContext(1), &pb.GetPlayerResourcesByCategoryIDRequest{
		PlayerID:           c.Player.GetID(),
		ResourceCategoryID: 3,
	}); err != nil {
		c.Logger.Panic(err)
	}

	return c.GetResourceList(rGetPlayerResource.GetPlayerInventory())
}

func (c *PlayerInventoryResourceController) GetResourceList(inventory []*pb.PlayerInventory) (recapResources string) {
	// Ordino l'array per rarità (dal più piccolo al più grande)
	resources := helpers.SortInventoryByRarity(inventory)

	for _, resource := range resources {
		if resource.GetQuantity() > 0 {
			recapResources += fmt.Sprintf(
				"- %s %v x %s (<b>%s</b>) %s\n",
				helpers.GetResourceCategoryIcons(resource.GetResource().GetResourceCategoryID()),
				resource.GetQuantity(),
				resource.GetResource().GetName(),
				strings.ToUpper(resource.GetResource().GetRarity().GetSlug()),
				helpers.GetResourceBaseIcons(resource.GetResource().GetBase()),
			)
		}
	}

	return
}
