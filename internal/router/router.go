package router

import (
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/internal/controllers"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
)

var (
	Routes = map[string]reflect.Type{
		"route.menu":              reflect.TypeOf((*controllers.MenuController)(nil)).Elem(),
		"route.tutorial":          reflect.TypeOf((*controllers.TutorialController)(nil)).Elem(),
		"route.tutorial.continue": reflect.TypeOf((*controllers.TutorialController)(nil)).Elem(),

		"route.exploration": reflect.TypeOf((*controllers.ExplorationController)(nil)).Elem(),
		"route.hunting":     reflect.TypeOf((*controllers.HuntingController)(nil)).Elem(),

		// Ship
		"route.ship":            reflect.TypeOf((*controllers.ShipController)(nil)).Elem(),
		"route.ship.travel":     reflect.TypeOf((*controllers.ShipTravelController)(nil)).Elem(),
		"route.ship.repairs":    reflect.TypeOf((*controllers.ShipRepairsController)(nil)).Elem(),
		"route.ship.rests":      reflect.TypeOf((*controllers.ShipRestsController)(nil)).Elem(),
		"route.ship.laboratory": reflect.TypeOf((*controllers.ShipLaboratoryController)(nil)).Elem(),

		// Player
		"route.player":              reflect.TypeOf((*controllers.PlayerController)(nil)).Elem(),
		"route.inventory":           reflect.TypeOf((*controllers.InventoryController)(nil)).Elem(),
		"route.inventory.resources": reflect.TypeOf((*controllers.InventoryResourceController)(nil)).Elem(),
		"route.inventory.items":     reflect.TypeOf((*controllers.InventoryItemController)(nil)).Elem(),
		"route.inventory.equip":     reflect.TypeOf((*controllers.PlayerEquipmentController)(nil)).Elem(),
		"route.banned":              reflect.TypeOf((*controllers.BannedController)(nil)).Elem(),

		// Planet
		"route.planet": reflect.TypeOf((*controllers.PlanetController)(nil)).Elem(),

		// Safe Planet
		"route.safeplanet.bank":                reflect.TypeOf((*controllers.SafePlanetBankController)(nil)).Elem(),
		"route.safeplanet.crafter":             reflect.TypeOf((*controllers.SafePlanetCrafterController)(nil)).Elem(),
		"route.safeplanet.coalition":           reflect.TypeOf((*controllers.SafePlanetCoalitionController)(nil)).Elem(),
		"route.safeplanet.mission":             reflect.TypeOf((*controllers.SafePlanetMissionController)(nil)).Elem(),
		"route.safeplanet.titan":               reflect.TypeOf((*controllers.SafePlanetTitanController)(nil)).Elem(),
		"route.safeplanet.coalition.expansion": reflect.TypeOf((*controllers.SafePlanetExpansionController)(nil)).Elem(),

		// Titan Planet
		"route.titanplanet.tackle": reflect.TypeOf((*controllers.TitanPlanetTackleController)(nil)).Elem(),

		// Conqueror
		"route.conqueror": reflect.TypeOf((*controllers.ConquerorController)(nil)).Elem(),
	}
)

// Routing - Effetua check sul tipo di messagio ed esegue un routing
func Routing(player *pb.Player, update tgbotapi.Update) {
	// A prescindere da tutto verifico se il player è stato bannato
	// Se così fosse non gestisco nemmeno l'update.
	if player.Banned {
		invoke(Routes["route.banned"], "Handle", player, update)
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
	isRoute, route := inRoutes(player.Language.Slug, callingRoute, Routes)
	if isRoute {
		invoke(Routes[route], "Handle", player, update)
		return
	}

	// Verifico se in memorià è presente già una rotta e se quella richiamata non sia menu
	// userò quella come main per gestire ulteriori sottostati
	isCachedRoute, _ := helpers.GetCurrentControllerCache(player.ID)
	if isCachedRoute != "" {
		invoke(Routes[isCachedRoute], "Handle", player, update)
		return
	}

	// Se nulla di tutto questo dovesse andare ritorno il menu
	invoke(Routes["route.menu"], "Handle", player, update)
}

// inRoutes - Verifica se esiste la rotta
func inRoutes(lang string, messageRoute string, routeList map[string]reflect.Type) (isRoute bool, route string) {
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
func invoke(any reflect.Type, name string, args ...interface{}) {
	// Recupero possibili input e li trasformo come argomenti da passare al metodo
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	v := reflect.New(any)
	v.MethodByName(name).Call(inputs)
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
