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
		"route.crafting":    new(controllers.CraftingV2Controller), // crafting.go  - MAIN
		"route.abilityTree": new(controllers.AbilityController),  // ability.go - MAIN

		"route.hunting": new(controllers.HuntingController), // hunting.go

		"route.inventory":         new(controllers.InventoryController),        // inventory.go - KEYBOARD
		"route.inventory.recap":   new(controllers.InventoryRecapController),   // inventory.go - MAIN
		"route.inventory.equip":   new(controllers.InventoryEquipController),   // inventory_equip.go - MAIN
		"route.inventory.destroy": new(controllers.InventoryDestroyController), // inventory_destroy.go - MAIN

		"route.ship":             new(controllers.ShipController),            // ship.go
		"route.ship.exploration": new(controllers.ShipExplorationController), // ship.go
		"route.ship.repairs":     new(controllers.ShipRepairsController),     // ship.go

		"route.testing.multiStage": new(controllers.TestingController),
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
