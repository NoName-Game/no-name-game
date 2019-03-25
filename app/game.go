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
		"the-answer-is":     controllers.TheAnswerIs,    // testing.go
		"test-multi-state":  controllers.TestMultiState, // testing.go
		"test-multi-stage":  controllers.TestMultiStage, // testing.go
		"back":              controllers.Back,           // back.go
		"clears":            controllers.Clears,         // clears.go
		"start":             controllers.StartTutorial,  // tutorial.go
		"time":              controllers.TestTimedQuest,
		"mission":           controllers.StartMission,
		"crafting":          controllers.Crafting,         // crafting.go
		"inventory":         controllers.Inventory,        // inventory.go
		"inventory-recap":   controllers.InventoryRecap,   // inventory.go - GLOBAL
		"equip":             controllers.InventoryEquip,   // inventory.go - GLOBAL FIXME: fix route name
		"inventory-destroy": controllers.InventoryDestroy, // inventory.go - GLOBAL
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
