package app

import (
	"reflect"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/app/controllers"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	Routes = map[string]reflect.Type{
		"route.menu":     reflect.TypeOf((*controllers.MenuController)(nil)).Elem(),
		"route.tutorial": reflect.TypeOf((*controllers.TutorialController)(nil)).Elem(),

		"route.exploration": reflect.TypeOf((*controllers.ExplorationController)(nil)).Elem(),
		"route.crafting":    reflect.TypeOf((*controllers.CraftingController)(nil)).Elem(),

		"route.hunting": reflect.TypeOf((*controllers.HuntingController)(nil)).Elem(),

		"route.ship":         reflect.TypeOf((*controllers.ShipController)(nil)).Elem(),
		"route.ship.travel":  reflect.TypeOf((*controllers.ShipTravelController)(nil)).Elem(),
		"route.ship.repairs": reflect.TypeOf((*controllers.ShipRepairsController)(nil)).Elem(),
		"route.ship.rests":   reflect.TypeOf((*controllers.ShipRestsController)(nil)).Elem(),

		// Player
		"route.player": reflect.TypeOf((*controllers.PlayerController)(nil)).Elem(),
		//"route.ability":         reflect.TypeOf((*controllers.AbilityController)(nil)).Elem(),
		"route.inventory":       reflect.TypeOf((*controllers.InventoryController)(nil)).Elem(),
		"route.inventory.recap": reflect.TypeOf((*controllers.InventoryRecapController)(nil)).Elem(),
		"route.inventory.equip": reflect.TypeOf((*controllers.PlayerEquipmentController)(nil)).Elem(),
		// "route.inventory.destroy": reflect.TypeOf((*controllers.InventoryDestroyController)(nil)).Elem(),
		"route.inventory.items": reflect.TypeOf((*controllers.InventoryItemController)(nil)).Elem(),
		"route.banned":          reflect.TypeOf((*controllers.BannedController)(nil)).Elem(),

		// Planet
		"route.planet": reflect.TypeOf((*controllers.PlanetController)(nil)).Elem(),

		// Safe Planet
		"route.safeplanet.bank":      reflect.TypeOf((*controllers.SafePlanetBankController)(nil)).Elem(),
		"route.safeplanet.crafter":   reflect.TypeOf((*controllers.SafePlanetCrafterController)(nil)).Elem(),
		"route.menu.npc":             reflect.TypeOf((*controllers.NpcMenuController)(nil)).Elem(), // TODO: da verificare
		"route.safeplanet.coalition": reflect.TypeOf((*controllers.SafePlanetCoalitionController)(nil)).Elem(),
		"route.safeplanet.mission":   reflect.TypeOf((*controllers.SafePlanetMissionController)(nil)).Elem(),
		"route.safeplanet.titan":     reflect.TypeOf((*controllers.SafePlanetTitanController)(nil)).Elem(),

		// Titan Planet
		"route.titanplanet.tackle": reflect.TypeOf((*controllers.TitanPlanetTackleController)(nil)).Elem(),
	}
)

// Init
func init() {
	// Inizializzo servizi bot
	var err = bootstrap()
	if err != nil {
		// Nel caso in cui uno dei servizi principale
		// dovesse entrare in errore in questo caso è meglio panicare
		panic(err)
	}
}

// Run - The Game!
func Run() {
	var err error

	// Recupero stati/messaggio da telegram
	updates, err := services.GetUpdates()
	if err != nil {
		services.ErrorHandler("Update channel error", err)
	}

	// Gestisco update ricevuti
	for update := range updates {
		// Gestisco singolo update in worker dedicato
		// go handleUpdate(update)
		handleUpdate(update)
	}
}

// handleUpdate - Gestisco singolo update
func handleUpdate(update tgbotapi.Update) {
	// Differisco controllo panic/recover
	// defer func() {
	// 	// Nel caso in cui panicasse
	// 	if err := recover(); err != nil {
	// 		// Registro errore
	// 		services.ErrorHandler("recover handle update", err.(error))
	// 	}
	// }()

	var err error
	// Gestisco utente
	var player *pb.Player
	player, err = helpers.HandleUser(update)
	if err != nil {
		panic(err)
	}

	// Gestisco update
	routing(player, update)
}
