package app

import (
	"bitbucket.org/no-name-game/nn-telegram/app/controllers"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

var (
	//===================================
	// Routes
	//
	Routes = map[string]interface{}{
		"route.start":    new(controllers.TutorialController), // tutorial.go  - MAIN
		"route.mission":  new(controllers.MissionController),  // mission.go   - MAIN
		"route.crafting": new(controllers.CraftingController), // crafting.go  - MAIN
		// "route.abilityTree": new(controllers.AbilityTree),   // ability.go - MAIN
		// "route.hunting":     new(controllers.Hunting),       // hunting.go
		// "route.menu":        new(controllers.Menu),          // menu.go

		// "route.inventory":         new(controllers.Inventory),        // inventory.go - KEYBOARD
		// "route.inventory.recap":   new(controllers.InventoryRecap),   // inventory.go - MAIN
		// "route.inventory.equip":   new(controllers.InventoryEquip),   // inventory.go - MAIN
		// "route.inventory.destroy": new(controllers.InventoryDestroy), // inventory.go - MAIN

		// "route.ship":             new(controllers.Ship),            // ship.go
		// "route.ship.exploration": new(controllers.ShipExploration), // ship.go

		// "route.ship.repairs": new(controllers.ShipRepairs), // ship.go

		// "route.testing.theAnswerIs": new(controllers.TheAnswerIs),    // testing.go
		// "route.testing.multiState":  new(controllers.TestMultiState), // testing.go
		// "route.testing.multiStage":  new(controllers.TestMultiStage), // testing.go
		// "route.testing.time":        new(controllers.TestTimedQuest), // testing.go

		// "callback.map": new(controllers.MapController),

		"route.testing.multiStageRevelution": new(controllers.TestingController),
	}

	BreakerRoutes = map[string]interface{}{
		"route.breaker.back":   controllers.Back,   // back.go      - MAIN (breaker)
		"route.breaker.clears": controllers.Clears, // clears.go    - MAIN (breaker)
	}
	//
	// End routes
	//=====================================
)

func init() {
	bootstrap()
}

// Run - The Game!
func Run() {
	updates, err := services.GetUpdates()
	if err != nil {
		services.ErrorHandler("Update channel error", err)
	}

	for update := range updates {
		// ***************
		// Handle users
		// ***************
		if !helpers.HandleUser(update) {
			continue
		}

		// ***************
		// Routing update
		// ***************
		routing(update)
	}
}
