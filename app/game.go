package app

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/controllers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

var (
	// ===================================
	// Routes
	//
	Routes = map[string]interface{}{
		// "route.menu":        new(controllers.MenuController),       // menu.go
		"route.start": new(controllers.TutorialController), // tutorial.go  - MAIN
		// "route.mission":     new(controllers.MissionController),    // mission.go   - MAIN
		// "route.crafting":    new(controllers.CraftingV2Controller), // crafting.go  - MAIN
		// "route.abilityTree": new(controllers.AbilityController),    // ability.go - MAIN
		//
		// "route.hunting": new(controllers.HuntingController), // hunting.go
		//
		// "route.inventory":         new(controllers.InventoryController),        // inventory.go - KEYBOARD
		// "route.inventory.recap":   new(controllers.InventoryRecapController),   // inventory.go - MAIN
		// "route.inventory.equip":   new(controllers.InventoryEquipController),   // inventory_equip.go - MAIN
		// "route.inventory.destroy": new(controllers.InventoryDestroyController), // inventory_destroy.go - MAIN
		//
		// "route.ship":             new(controllers.ShipController),            // ship.go
		// "route.ship.exploration": new(controllers.ShipExplorationController), // ship.go
		// "route.ship.repairs":     new(controllers.ShipRepairsController),     // ship.go
		//
		// "route.testing.multiStage": new(controllers.TestingController),
	}

	BreakerRoutes = map[string]interface{}{
		// "route.breaker.back":   new(controllers.BackController),   // breaker.go      - MAIN (breaker)
		// "route.breaker.clears": new(controllers.ClearsController), // breaker.go    - MAIN (breaker)
	}
	//
	// End routes
	// =====================================
)

// Init
func init() {
	// Inizializzo servizi bot
	var err error
	err = bootstrap()
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
	// // Mando un messaggio dicendogli di inserire un username
	// msg := services.NewMessage(update.Message.Chat.ID, Trans("en", "miss_username"))
	// services.SendMessage(msg)
	// 	}
	// }()

	var err error
	// Gestisco utente
	var player nnsdk.Player
	player, err = helpers.HandleUser(update)
	if err != nil {
		panic(err)
	}

	// Gestisco update
	routing(player, update)

	// Se è un player già registrato verifico che non sia morto per continuare
	// if player.Stats.Dead == true {
	// 	// controllers.PlayerDeath(update) TODO: FIXME
	// 	return false
	// }
}
