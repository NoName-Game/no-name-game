package router

import (
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/internal/controllers"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
)

var routes = map[string]controllers.ControllerInterface{
	"route.menu":              &controllers.MenuController{},
	"route.tutorial":          &controllers.TutorialController{},
	"route.tutorial.continue": &controllers.TutorialController{},

	"route.exploration": &controllers.ExplorationController{},
	"route.hunting":     &controllers.HuntingController{},

	// Ship
	"route.ship":            &controllers.ShipController{},
	"route.ship.travel":     &controllers.ShipTravelController{},
	"route.ship.repairs":    &controllers.ShipRepairsController{},
	"route.ship.rests":      &controllers.ShipRestsController{},
	"route.ship.laboratory": &controllers.ShipLaboratoryController{},

	// Player
	"route.player":              &controllers.PlayerController{},
	"route.inventory":           &controllers.InventoryController{},
	"route.inventory.resources": &controllers.InventoryResourceController{},
	"route.inventory.items":     &controllers.InventoryItemController{},
	"route.inventory.equip":     &controllers.PlayerEquipmentController{},
	"route.banned":              &controllers.BannedController{},

	// Planet
	"route.planet": &controllers.PlanetController{},

	// Safe Planet
	"route.safeplanet.bank":                &controllers.SafePlanetBankController{},
	"route.safeplanet.crafter":             &controllers.SafePlanetCrafterController{},
	"route.safeplanet.coalition":           &controllers.SafePlanetCoalitionController{},
	"route.safeplanet.mission":             &controllers.SafePlanetMissionController{},
	"route.safeplanet.titan":               &controllers.SafePlanetTitanController{},
	"route.safeplanet.coalition.expansion": &controllers.SafePlanetExpansionController{},

	// Titan Planet
	"route.titanplanet.tackle": &controllers.TitanPlanetTackleController{},

	// Conqueror
	"route.conqueror": &controllers.ConquerorController{},
}

// Routing - Effetua check sul tipo di messagio ed esegue un routing
func Routing(player *pb.Player, update tgbotapi.Update) {
	// A prescindere da tutto verifico se il player è stato bannato
	// Se così fosse non gestisco nemmeno l'update.
	if player.Banned {
		invoke(routes["route.banned"], player, update)
		return
	}

	// Verifica il tipo di messaggio
	var callingRoute string
	if update.Message != nil {
		callingRoute = parseMessage(update.Message)
	} else if update.CallbackQuery != nil {
		callingRoute = parseCallback(update.CallbackQuery)
	}

	// Dirigo ad una rotta normale
	isRoute, route := inRoutes(player.Language.Slug, callingRoute, routes)
	if isRoute {
		invoke(routes[route], player, update)
		return
	}

	// Verifico se in memorià è presente già una rotta e se quella richiamata non sia menu
	// userò quella come main per gestire ulteriori sottostati
	cachedRoute, _ := helpers.GetCurrentControllerCache(player.ID)
	if cachedRoute != "" {
		invoke(routes[cachedRoute], player, update)
		return
	}

	// Se nulla di tutto questo dovesse andare ritorno il menu
	invoke(routes["route.menu"], player, update)
}

// inRoutes - Verifica se esiste la rotta
func inRoutes(lang string, messageRoute string, routeList map[string]controllers.ControllerInterface) (isRoute bool, route string) {
	// Ciclo lista di rotte
	for route := range routeList {
		// Traduco le rotte in base alla lingua del player per trovare corrispondenza
		if strings.ToLower(helpers.Trans(lang, route)) == messageRoute {
			return true, route
		}
	}

	return false, ""
}

// invoke - Invoco dinamicamente un metodo di un controller
func invoke(c controllers.ControllerInterface, player *pb.Player, update tgbotapi.Update) {
	c.Handle(player, update)
}

// Metodo per il parsing del messaggio
func parseMessage(message *tgbotapi.Message) (parsed string) {
	parsed = message.Text
	if message.IsCommand() {
		parsed = message.Command()
		// Se è un comando ed è start lo parso come tutorial
		if parsed == "start" {
			parsed = "📖 Tutorial"
		}
	}

	return strings.ToLower(parsed)
}

// Metodo per il parsing della callback
func parseCallback(callback *tgbotapi.CallbackQuery) (parsed string) {
	// Prendo la prima parte del callback che contiene la rotta
	parsed = strings.Split(callback.Data, ".")[0]

	return strings.ToLower(parsed)
}
