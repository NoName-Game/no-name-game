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
		"the-answer-is":     controllers.TheAnswerIs,      // testing.go
		"test-multi-state":  controllers.TestMultiState,   // testing.go
		"test-multi-stage":  controllers.TestMultiStage,   // testing.go
		"back":              controllers.Back,             // back.go      - MAIN (breaker)
		"clears":            controllers.Clears,           // clears.go    - MAIN (breaker)
		"start":             controllers.StartTutorial,    // tutorial.go  - MAIN
		"time":              controllers.TestTimedQuest,   //
		"mission":           controllers.StartMission,     // mission.go   - MAIN
		"crafting":          controllers.Crafting,         // crafting.go  - MAIN
		"inventory":         controllers.Inventory,        // inventory.go - KEYBOARD
		"inventory-recap":   controllers.InventoryRecap,   // inventory.go - MAIN
		"inventory-equip":   controllers.InventoryEquip,   // inventory.go - MAIN
		"inventory-destroy": controllers.InventoryDestroy, // inventory.go - MAIN
	}

	breakerRoutes = []string{
		"back",   // back.go
		"clears", // clears.go
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
