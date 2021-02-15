package router

import (
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"

	"bitbucket.org/no-name-game/nn-telegram/internal/controllers"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
)

var routes = map[string]reflect.Type{
	"route.menu":              reflect.TypeOf((*controllers.MenuController)(nil)).Elem(),
	"route.setup":             reflect.TypeOf((*controllers.SetupController)(nil)).Elem(),
	"route.tutorial":          reflect.TypeOf((*controllers.TutorialController)(nil)).Elem(),
	"route.tutorial.continue": reflect.TypeOf((*controllers.TutorialController)(nil)).Elem(),

	"route.exploration":  reflect.TypeOf((*controllers.ExplorationController)(nil)).Elem(),
	"route.hunting":      reflect.TypeOf((*controllers.HuntingController)(nil)).Elem(),
	"route.darkmerchant": reflect.TypeOf((*controllers.DarkMerchantController)(nil)).Elem(),

	// Ship
	"route.ship":                 reflect.TypeOf((*controllers.ShipController)(nil)).Elem(),
	"route.ship.travel":          reflect.TypeOf((*controllers.ShipTravelController)(nil)).Elem(),
	"route.ship.travel.finding":  reflect.TypeOf((*controllers.ShipTravelFindingController)(nil)).Elem(),
	"route.ship.travel.manual":   reflect.TypeOf((*controllers.ShipTravelManualController)(nil)).Elem(),
	"route.ship.travel.rescue":   reflect.TypeOf((*controllers.ShipTravelRescueController)(nil)).Elem(),
	"ship.travel.land":           reflect.TypeOf((*controllers.ShipTravelFindingController)(nil)).Elem(),
	"route.ship.travel.favorite": reflect.TypeOf((*controllers.ShipTravelFindingController)(nil)).Elem(),
	"route.ship.rests":           reflect.TypeOf((*controllers.ShipRestsController)(nil)).Elem(),
	"ship.rests.wakeup":          reflect.TypeOf((*controllers.ShipRestsController)(nil)).Elem(),
	"route.ship.laboratory":      reflect.TypeOf((*controllers.ShipLaboratoryController)(nil)).Elem(),

	// Player
	"route.player":                     reflect.TypeOf((*controllers.PlayerController)(nil)).Elem(),
	"route.player.guild":               reflect.TypeOf((*controllers.PlayerGuildController)(nil)).Elem(),
	"route.player.party":               reflect.TypeOf((*controllers.PlayerPartyController)(nil)).Elem(),
	"route.player.party.create":        reflect.TypeOf((*controllers.PlayerPartyCreateController)(nil)).Elem(),
	"route.player.party.leave":         reflect.TypeOf((*controllers.PlayerPartyLeaveController)(nil)).Elem(),
	"route.player.party.add_player":    reflect.TypeOf((*controllers.PlayerPartyAddPlayerController)(nil)).Elem(),
	"route.player.party.remove_player": reflect.TypeOf((*controllers.PlayerPartyRemovePlayerController)(nil)).Elem(),
	"route.player.achievements":        reflect.TypeOf((*controllers.PlayerAchievementsController)(nil)).Elem(),
	"route.player.inventory":           reflect.TypeOf((*controllers.PlayerInventoryController)(nil)).Elem(),
	"route.player.inventory.resources": reflect.TypeOf((*controllers.PlayerInventoryResourceController)(nil)).Elem(),
	"route.player.inventory.items":     reflect.TypeOf((*controllers.PlayerInventoryItemController)(nil)).Elem(),
	"route.player.inventory.equip":     reflect.TypeOf((*controllers.PlayerEquipmentController)(nil)).Elem(),
	"route.banned":                     reflect.TypeOf((*controllers.BannedController)(nil)).Elem(),

	// Planet
	"route.planet":                 reflect.TypeOf((*controllers.PlanetController)(nil)).Elem(),
	"route.planet.bookmark.add":    reflect.TypeOf((*controllers.PlanetBookmarkAddController)(nil)).Elem(),
	"route.planet.bookmark.remove": reflect.TypeOf((*controllers.PlanetBookmarkRemoveController)(nil)).Elem(),

	// Safe Planet
	"route.safeplanet.coalition.daily_reward":             reflect.TypeOf((*controllers.SafePlanetCoalitionDailyRewardController)(nil)).Elem(),
	"route.safeplanet.bank":                               reflect.TypeOf((*controllers.SafePlanetBankController)(nil)).Elem(),
	"route.safeplanet.crafter":                            reflect.TypeOf((*controllers.SafePlanetCrafterController)(nil)).Elem(),
	"route.safeplanet.crafter.create":                     reflect.TypeOf((*controllers.SafePlanetCrafterCreateController)(nil)).Elem(),
	"route.safeplanet.crafter.repair":                     reflect.TypeOf((*controllers.SafePlanetCrafterRepairController)(nil)).Elem(),
	"route.safeplanet.crafter.decompose":                  reflect.TypeOf((*controllers.SafePlanetCrafterDecomposeController)(nil)).Elem(),
	"route.safeplanet.dealer":                             reflect.TypeOf((*controllers.SafePlanetDealerController)(nil)).Elem(),
	"route.safeplanet.accademy":                           reflect.TypeOf((*controllers.SafePlanetAccademyController)(nil)).Elem(),
	"route.safeplanet.relax":                              reflect.TypeOf((*controllers.SafePlanetRelaxController)(nil)).Elem(),
	"route.safeplanet.relax.threecard":                    reflect.TypeOf((*controllers.SafePlanetRelaxThreeCardController)(nil)).Elem(),
	"route.safeplanet.hangar":                             reflect.TypeOf((*controllers.SafePlanetHangarController)(nil)).Elem(),
	"route.safeplanet.hangar.ships":                       reflect.TypeOf((*controllers.SafePlanetHangarShipsController)(nil)).Elem(),
	"route.safeplanet.hangar.repair":                      reflect.TypeOf((*controllers.SafePlanetHangarRepairController)(nil)).Elem(),
	"route.safeplanet.hangar.create":                      reflect.TypeOf((*controllers.SafePlanetHangarCreateController)(nil)).Elem(),
	"route.safeplanet.coalition":                          reflect.TypeOf((*controllers.SafePlanetCoalitionController)(nil)).Elem(),
	"route.safeplanet.coalition.mission":                  reflect.TypeOf((*controllers.SafePlanetMissionController)(nil)).Elem(),
	"route.safeplanet.coalition.titan":                    reflect.TypeOf((*controllers.SafePlanetTitanController)(nil)).Elem(),
	"route.safeplanet.coalition.expansion":                reflect.TypeOf((*controllers.SafePlanetExpansionController)(nil)).Elem(),
	"route.safeplanet.coalition.research":                 reflect.TypeOf((*controllers.SafePlanetResearchController)(nil)).Elem(),
	"route.safeplanet.coalition.research.donation":        reflect.TypeOf((*controllers.SafePlanetResearchDonationController)(nil)).Elem(),
	"route.safeplanet.coalition.statistics":               reflect.TypeOf((*controllers.SafePlanetCoalitionStatisticsController)(nil)).Elem(),
	"route.safeplanet.coalition.protectors":               reflect.TypeOf((*controllers.SafePlanetProtectorsController)(nil)).Elem(),
	"route.safeplanet.coalition.protectors.create":        reflect.TypeOf((*controllers.SafePlanetProtectorsCreateController)(nil)).Elem(),
	"route.safeplanet.coalition.protectors.join":          reflect.TypeOf((*controllers.SafePlanetProtectorsJoinController)(nil)).Elem(),
	"route.safeplanet.coalition.protectors.leave":         reflect.TypeOf((*controllers.SafePlanetProtectorsLeaveController)(nil)).Elem(),
	"route.safeplanet.coalition.protectors.add_player":    reflect.TypeOf((*controllers.SafePlanetProtectorsAddPlayerController)(nil)).Elem(),
	"route.safeplanet.coalition.protectors.remove_player": reflect.TypeOf((*controllers.SafePlanetProtectorsRemovePlayerController)(nil)).Elem(),

	// Titan Planet
	"route.titanplanet.tackle": reflect.TypeOf((*controllers.TitanPlanetTackleController)(nil)).Elem(),

	// Conqueror
	"route.conqueror": reflect.TypeOf((*controllers.ConquerorController)(nil)).Elem(),
}

// Routing - Effetua check sul tipo di messagio ed esegue un routing
func Routing(player *pb.Player, update tgbotapi.Update) {
	// A prescindere da tutto verifico se il player è stato bannato
	// Se così fosse non gestisco nemmeno l'update.
	if player.Banned {
		invoke(routes["route.banned"], "Handle", player, update)
		return
	}

	// Verifica il tipo di messaggio
	var callingRoute string
	if update.Message != nil {
		callingRoute = parseMessage(update.Message)
	} else if update.CallbackQuery != nil {
		callingRoute = parseCallback(update.CallbackQuery)
	}

	// Se morto spedisco direttamente al riposo
	// è necessario effettuare il controllo di hunting in quanto è necessario
	//  per cancellare le attività in corso
	if callingRoute != "hunting" {
		if player.Dead {
			invoke(routes["route.ship.rests"], "Handle", player, update)
			return
		}
	}

	// Dirigo ad una rotta normale
	isRoute, route := inRoutes(player.Language.Slug, callingRoute, routes)
	if isRoute {
		invoke(routes[route], "Handle", player, update)
		return
	}

	// Verifico se in memorià è presente già una rotta e se quella richiamata non sia menu
	// userò quella come main per gestire ulteriori sottostati
	var cachedRoute string
	// Non è necessario verificare l'errore perchè non per forza deve eserci una rotta in cache
	cachedRoute, _ = helpers.GetCurrentControllerCache(player.ID)
	if cachedRoute != "" {
		if _, ok := routes[cachedRoute]; ok {
			invoke(routes[cachedRoute], "Handle", player, update)
			return
		}

		logrus.Errorf("invalid cached route value: %s", cachedRoute)
	}

	// Se nulla di tutto questo dovesse andare ritorno il menu
	invoke(routes["route.menu"], "Handle", player, update)
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
func invoke(controller reflect.Type, method string, args ...interface{}) {
	// Recupero possibili input e li trasformo come argomenti da passare al metodo
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}

	v := reflect.New(controller)
	v.MethodByName(method).Call(inputs)
}

// Metodo per il parsing del messaggio
func parseMessage(message *tgbotapi.Message) (parsed string) {
	parsed = message.Text
	if message.IsCommand() {
		parsed = message.Command()
		// Se è un comando ed è start lo parso come tutorial
		if parsed == "start" {
			parsed = "⚙️ Setup"
		}
	}

	return strings.ToLower(parsed)
}

// Metodo per il parsing della callback
func parseCallback(callback *tgbotapi.CallbackQuery) (parsed string) {
	// Recupero infomazioni callback
	var inlineData helpers.InlineDataStruct
	inlineData = inlineData.GetDataValue(callback.Data)

	return strings.ToLower(inlineData.C)
}
