package router

import (
	"reflect"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/internal/controllers"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
)

type Route struct {
	Route      string
	Controller controllers.ControllerInterface
	Reflect    reflect.Type
}

var router = []Route{
	{Route: "route.menu", Controller: &controllers.MenuController{}, Reflect: reflect.TypeOf((*controllers.MenuController)(nil)).Elem()},
	{Route: "route.setup", Controller: &controllers.SetupController{}, Reflect: reflect.TypeOf((*controllers.SetupController)(nil)).Elem()},
	{Route: "route.info", Controller: &controllers.InfoController{}, Reflect: reflect.TypeOf((*controllers.InfoController)(nil)).Elem()},
	{Route: "route.banned", Controller: &controllers.BannedController{}, Reflect: reflect.TypeOf((*controllers.BannedController)(nil)).Elem()},
	{Route: "route.tutorial", Controller: &controllers.TutorialController{}, Reflect: reflect.TypeOf((*controllers.TutorialController)(nil)).Elem()},
	{Route: "route.tutorial.continue", Controller: &controllers.TutorialController{}, Reflect: reflect.TypeOf((*controllers.TutorialController)(nil)).Elem()},

	{Route: "route.exploration", Controller: &controllers.ExplorationController{}, Reflect: reflect.TypeOf((*controllers.ExplorationController)(nil)).Elem()},
	{Route: "route.hunting", Controller: &controllers.HuntingController{}, Reflect: reflect.TypeOf((*controllers.HuntingController)(nil)).Elem()},
	{Route: "route.darkmerchant", Controller: &controllers.DarkMerchantController{}, Reflect: reflect.TypeOf((*controllers.DarkMerchantController)(nil)).Elem()},

	// Ship
	{Route: "route.ship", Controller: &controllers.ShipController{}, Reflect: reflect.TypeOf((*controllers.ShipController)(nil)).Elem()},
	{Route: "route.ship.travel", Controller: &controllers.ShipTravelController{}, Reflect: reflect.TypeOf((*controllers.ShipTravelController)(nil)).Elem()},
	{Route: "route.ship.travel.finding", Controller: &controllers.ShipTravelFindingController{}, Reflect: reflect.TypeOf((*controllers.ShipTravelFindingController)(nil)).Elem()},
	{Route: "route.ship.travel.manual", Controller: &controllers.ShipTravelManualController{}, Reflect: reflect.TypeOf((*controllers.ShipTravelManualController)(nil)).Elem()},
	{Route: "route.ship.travel.rescue", Controller: &controllers.ShipTravelRescueController{}, Reflect: reflect.TypeOf((*controllers.ShipTravelRescueController)(nil)).Elem()},
	{Route: "ship.travel.land", Controller: &controllers.ShipTravelFindingController{}, Reflect: reflect.TypeOf((*controllers.ShipTravelFindingController)(nil)).Elem()},
	{Route: "route.ship.travel.favorite", Controller: &controllers.ShipTravelFindingController{}, Reflect: reflect.TypeOf((*controllers.ShipTravelFindingController)(nil)).Elem()},
	{Route: "route.ship.rests", Controller: &controllers.ShipRestsController{}, Reflect: reflect.TypeOf((*controllers.ShipRestsController)(nil)).Elem()},
	{Route: "ship.rests.wakeup", Controller: &controllers.ShipRestsController{}, Reflect: reflect.TypeOf((*controllers.ShipRestsController)(nil)).Elem()},
	{Route: "route.ship.laboratory", Controller: &controllers.ShipLaboratoryController{}, Reflect: reflect.TypeOf((*controllers.ShipLaboratoryController)(nil)).Elem()},

	// Player
	{Route: "route.player", Controller: &controllers.PlayerController{}, Reflect: reflect.TypeOf((*controllers.PlayerController)(nil)).Elem()},
	{Route: "route.player.guild", Controller: &controllers.PlayerGuildController{}, Reflect: reflect.TypeOf((*controllers.PlayerGuildController)(nil)).Elem()},
	{Route: "route.player.party", Controller: &controllers.PlayerPartyController{}, Reflect: reflect.TypeOf((*controllers.PlayerPartyController)(nil)).Elem()},
	{Route: "route.player.party.create", Controller: &controllers.PlayerPartyCreateController{}, Reflect: reflect.TypeOf((*controllers.PlayerPartyCreateController)(nil)).Elem()},
	{Route: "route.player.party.leave", Controller: &controllers.PlayerPartyLeaveController{}, Reflect: reflect.TypeOf((*controllers.PlayerPartyLeaveController)(nil)).Elem()},
	{Route: "route.player.party.add_player", Controller: &controllers.PlayerPartyAddPlayerController{}, Reflect: reflect.TypeOf((*controllers.PlayerPartyAddPlayerController)(nil)).Elem()},
	{Route: "route.player.party.remove_player", Controller: &controllers.PlayerPartyRemovePlayerController{}, Reflect: reflect.TypeOf((*controllers.PlayerPartyRemovePlayerController)(nil)).Elem()},
	{Route: "route.player.achievements", Controller: &controllers.PlayerAchievementsController{}, Reflect: reflect.TypeOf((*controllers.PlayerAchievementsController)(nil)).Elem()},
	{Route: "route.player.inventory", Controller: &controllers.PlayerInventoryController{}, Reflect: reflect.TypeOf((*controllers.PlayerInventoryController)(nil)).Elem()},
	{Route: "route.player.inventory.resources", Controller: &controllers.PlayerInventoryResourceController{}, Reflect: reflect.TypeOf((*controllers.PlayerInventoryResourceController)(nil)).Elem()},
	{Route: "route.player.inventory.items", Controller: &controllers.PlayerInventoryItemController{}, Reflect: reflect.TypeOf((*controllers.PlayerInventoryItemController)(nil)).Elem()},
	{Route: "route.player.inventory.packs", Controller: &controllers.PlayerInventoryPackController{}, Reflect: reflect.TypeOf((*controllers.PlayerInventoryPackController)(nil)).Elem()},
	{Route: "route.player.inventory.equip", Controller: &controllers.PlayerEquipmentController{}, Reflect: reflect.TypeOf((*controllers.PlayerEquipmentController)(nil)).Elem()},

	// Planet
	{Route: "route.planet", Controller: &controllers.PlanetController{}, Reflect: reflect.TypeOf((*controllers.PlanetController)(nil)).Elem()},
	{Route: "route.planet.bookmark.add", Controller: &controllers.PlanetBookmarkAddController{}, Reflect: reflect.TypeOf((*controllers.PlanetBookmarkAddController)(nil)).Elem()},
	{Route: "route.planet.bookmark.remove", Controller: &controllers.PlanetBookmarkRemoveController{}, Reflect: reflect.TypeOf((*controllers.PlanetBookmarkRemoveController)(nil)).Elem()},
	{Route: "route.planet.conqueror", Controller: &controllers.PlanetConquerorController{}, Reflect: reflect.TypeOf((*controllers.PlanetConquerorController)(nil)).Elem()},
	{Route: "route.planet.domain", Controller: &controllers.PlanetDomainController{}, Reflect: reflect.TypeOf((*controllers.PlanetDomainController)(nil)).Elem()},

	// Safe Planet
	{Route: "route.safeplanet.coalition.daily_reward", Controller: &controllers.SafePlanetCoalitionDailyRewardController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetCoalitionDailyRewardController)(nil)).Elem()},
	{Route: "route.safeplanet.bank", Controller: &controllers.SafePlanetBankController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetBankController)(nil)).Elem()},
	{Route: "route.safeplanet.crafter", Controller: &controllers.SafePlanetCrafterController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetCrafterController)(nil)).Elem()},
	{Route: "route.safeplanet.crafter.create", Controller: &controllers.SafePlanetCrafterCreateController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetCrafterCreateController)(nil)).Elem()},
	{Route: "route.safeplanet.crafter.repair", Controller: &controllers.SafePlanetCrafterRepairController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetCrafterRepairController)(nil)).Elem()},
	{Route: "route.safeplanet.crafter.decompose", Controller: &controllers.SafePlanetCrafterDecomposeController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetCrafterDecomposeController)(nil)).Elem()},

	{Route: "route.safeplanet.market", Controller: &controllers.SafePlanetMarketController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetMarketController)(nil)).Elem()},
	{Route: "route.safeplanet.market.dealer", Controller: &controllers.SafePlanetMarketDealerController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetMarketDealerController)(nil)).Elem()},
	{Route: "route.safeplanet.market.gift", Controller: &controllers.SafePlanetMarketGiftController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetMarketGiftController)(nil)).Elem()},
	{Route: "route.safeplanet.market.shareholder", Controller: &controllers.SafePlanetMarketShareHolderController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetMarketShareHolderController)(nil)).Elem()},
	{Route: "route.safeplanet.market.auctions", Controller: &controllers.SafePlanetMarketAuctionsController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetMarketAuctionsController)(nil)).Elem()},
	{Route: "route.safeplanet.market.auctions.sell", Controller: &controllers.SafePlanetMarketAuctionsSellController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetMarketAuctionsSellController)(nil)).Elem()},
	{Route: "route.safeplanet.market.auctions.buy", Controller: &controllers.SafePlanetMarketAuctionsBuyController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetMarketAuctionsBuyController)(nil)).Elem()},
	{Route: "route.safeplanet.market.auctions.my_auction", Controller: &controllers.SafePlanetMarketAuctionsMyAuctionController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetMarketAuctionsMyAuctionController)(nil)).Elem()},

	{Route: "route.safeplanet.accademy", Controller: &controllers.SafePlanetAccademyController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetAccademyController)(nil)).Elem()},
	{Route: "route.safeplanet.relax", Controller: &controllers.SafePlanetRelaxController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetRelaxController)(nil)).Elem()},
	{Route: "route.safeplanet.relax.threecard", Controller: &controllers.SafePlanetRelaxThreeCardController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetRelaxThreeCardController)(nil)).Elem()},

	{Route: "route.safeplanet.hangar", Controller: &controllers.SafePlanetHangarController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetHangarController)(nil)).Elem()},
	{Route: "route.safeplanet.hangar.ships", Controller: &controllers.SafePlanetHangarShipsController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetHangarShipsController)(nil)).Elem()},
	{Route: "route.safeplanet.hangar.repair", Controller: &controllers.SafePlanetHangarRepairController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetHangarRepairController)(nil)).Elem()},
	{Route: "route.safeplanet.hangar.create", Controller: &controllers.SafePlanetHangarCreateController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetHangarCreateController)(nil)).Elem()},

	{Route: "route.safeplanet.coalition", Controller: &controllers.SafePlanetCoalitionController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetCoalitionController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.mission", Controller: &controllers.SafePlanetMissionController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetMissionController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.titan", Controller: &controllers.SafePlanetTitanController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetTitanController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.expansion", Controller: &controllers.SafePlanetExpansionController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetExpansionController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.research", Controller: &controllers.SafePlanetResearchController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetResearchController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.research.donation", Controller: &controllers.SafePlanetResearchDonationController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetResearchDonationController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.statistics", Controller: &controllers.SafePlanetCoalitionStatisticsController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetCoalitionStatisticsController)(nil)).Elem()},

	{Route: "route.safeplanet.coalition.protectors", Controller: &controllers.SafePlanetProtectorsController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetProtectorsController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.protectors.create", Controller: &controllers.SafePlanetProtectorsCreateController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetProtectorsCreateController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.protectors.join", Controller: &controllers.SafePlanetProtectorsJoinController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetProtectorsJoinController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.protectors.leave", Controller: &controllers.SafePlanetProtectorsLeaveController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetProtectorsLeaveController)(nil)).Elem()},
	// {Route: "route.safeplanet.coalition.protectors.add_player", Controller: &controllers.SafePlanetProtectorsAddPlayerController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetProtectorsAddPlayerController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.protectors.change_leader", Controller: &controllers.SafePlanetProtectorsChangeLeaderController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetProtectorsChangeLeaderController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.protectors.remove_player", Controller: &controllers.SafePlanetProtectorsRemovePlayerController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetProtectorsRemovePlayerController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.protectors.switch", Controller: &controllers.SafePlanetProtectorsSwitchController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetProtectorsSwitchController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.protectors.change_name", Controller: &controllers.SafePlanetProtectorsChangeNameController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetProtectorsChangeNameController)(nil)).Elem()},
	{Route: "route.safeplanet.coalition.protectors.change_tag", Controller: &controllers.SafePlanetProtectorsChangeTagController{}, Reflect: reflect.TypeOf((*controllers.SafePlanetProtectorsChangeTagController)(nil)).Elem()},

	// Titan Planet
	{Route: "route.titanplanet.tackle", Controller: &controllers.TitanPlanetTackleController{}, Reflect: reflect.TypeOf((*controllers.TitanPlanetTackleController)(nil)).Elem()},
}

// Routing - Effetua check sul tipo di messagio ed esegue un routing
func Routing(player *pb.Player, update tgbotapi.Update) {
	// A prescindere da tutto verifico se il player è stato bannato
	// Se così fosse non gestisco nemmeno l'update.
	if player.Banned {
		invoke(getRoute("route.banned").Reflect, "Handle", player, update)
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
	if callingRoute != "hunting" && callingRoute != strings.ToLower(helpers.Trans(player.GetLanguage().GetSlug(), "hunting.leave")) {
		if player.Dead {
			invoke(getRoute("route.ship.rests").Reflect, "Handle", player, update)
			return
		}
	}

	// Verifico se in memorià è presente già una rotta e se quella richiamata non sia menu
	// userò quella come main per gestire ulteriori sottostati
	var cachedRouteString string
	// Non è necessario verificare l'errore perchè non per forza deve eserci una rotta in cache
	cachedRouteString, _ = helpers.GetCurrentControllerCache(player.ID)
	if cachedRouteString != "" {
		// Recupero route da cached
		cachedRoute := getRoute(cachedRouteString)

		// Verifico se viene richiamato un nuovo metodo
		isRoute2, calledRoute := isRoutes(player.Language.Slug, callingRoute)

		// Verifico se nella cached route è permesso chimare il nuovo metodo
		if cachedRoute.Route != "" && isRoute2 && cachedRoute.Route != calledRoute.Route {
			// Recupero configurazione controller
			config := cachedRoute.Controller.Configuration(player, update)
			for _, allowed := range config.Configurations.AllowedControllers {
				if allowed == calledRoute.Route {
					invoke(calledRoute.Reflect, "Handle", player, update)
				}
			}

			// Se non è una rotta consentita escludo il messaggio
			return
		}

		// Chiamo rotta in cache
		invoke(cachedRoute.Reflect, "Handle", player, update)
		return
	}

	// Dirigo ad una rotta normale
	isRoute, route := isRoutes(player.Language.Slug, callingRoute)
	if isRoute {
		invoke(route.Reflect, "Handle", player, update)
		return
	}

	// Se nulla di tutto questo dovesse andare ritorno il menu
	invoke(getRoute("route.menu").Reflect, "Handle", player, update)
}

func getRoute(routeCalled string) Route {
	for _, route := range router {
		if route.Route == routeCalled {
			return route
		}
	}

	return Route{}
}

func isRoutes(lang string, routeCalled string) (bool, Route) {
	for _, route := range router {
		if strings.ToLower(helpers.Trans(lang, route.Route)) == routeCalled {
			return true, route
		}
	}

	return false, Route{}
}

// inRoutes - Verifica se esiste la rotta
// func inRoutes(lang string, messageRoute string, routeList map[string]reflect.Type) (isRoute bool, route string) {
// 	// Ciclo lista di rotte
// 	for route := range routeList {
// 		// Traduco le rotte in base alla lingua del player per trovare corrispondenza
// 		if strings.ToLower(helpers.Trans(lang, route)) == messageRoute {
// 			return true, route
// 		}
// 	}
//
// 	return false, ""
// }

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
