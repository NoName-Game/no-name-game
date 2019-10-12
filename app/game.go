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
		"route.menu":        new(controllers.MenuController),     // menu.go
		"route.start":       new(controllers.TutorialController), // tutorial.go  - MAIN
		"route.mission":     new(controllers.MissionController),  // mission.go   - MAIN
		"route.crafting":    new(controllers.CraftingController), // crafting.go  - MAIN
		"route.abilityTree": new(controllers.AbilityController),  // ability.go - MAIN

		"route.hunting": new(controllers.HuntingController), // hunting.go

		"route.inventory":       new(controllers.InventoryController),      // inventory.go - KEYBOARD
		"route.inventory.recap": new(controllers.InventoryRecapController), // inventory.go - MAIN
		"route.inventory.equip": new(controllers.InventoryEquipController), // inventory.go - MAIN
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
		"route.breaker.back":   new(controllers.BackController),   // breaker.go      - MAIN (breaker)
		"route.breaker.clears": new(controllers.ClearsController), // breaker.go    - MAIN (breaker)
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
