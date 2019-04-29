package app

import (
	"bitbucket.org/no-name-game/no-name/app/controllers"
	"bitbucket.org/no-name-game/no-name/app/helpers"
	"bitbucket.org/no-name-game/no-name/services"
)

var (
	//===================================
	// Routes
	//
	routes = map[string]interface{}{
		// "route.start":             controllers.StartTutorial,    // tutorial.go  - MAIN
		// "route.mission":           controllers.StartMission,     // mission.go   - MAIN
		"route.crafting":          controllers.Crafting,         // crafting.go  - MAIN
		"route.inventory":         controllers.Inventory,        // inventory.go - KEYBOARD
		"route.inventory.recap":   controllers.InventoryRecap,   // inventory.go - MAIN
		"route.inventory.equip":   controllers.InventoryEquip,   // inventory.go - MAIN
		"route.inventory.destroy": controllers.InventoryDestroy, // inventory.go - MAIN
		"route.abilityTree":       controllers.AbilityTree,      // ability.go - MAIN

		"route.testing.theAnswerIs": controllers.TheAnswerIs,    // testing.go
		"route.testing.multiState":  controllers.TestMultiState, // testing.go
		"route.testing.multiStage":  controllers.TestMultiStage, // testing.go
		"route.testing.time":        controllers.TestTimedQuest, // testing.go
	}

	breakerRoutes = map[string]interface{}{
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
	updates := services.GetUpdates()
	for update := range updates {
		if update.Message != nil {
			if update.Message.From.UserName == "" {
				msg := services.NewMessage(update.Message.Chat.ID, helpers.Trans("miss_username", "en"))
				services.SendMessage(msg)
				continue
			}

			routing(update)
		}
	}
}
