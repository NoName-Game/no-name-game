package app

import (
	"bitbucket.org/no-name-game/no-name/services"
)

var (
	player Player

	//===================================
	// Routes
	routes = map[string]interface{}{
		"the-answer-is":    theAnswerIs,
		"test-multi-state": testMultiState,
		"test-multi-stage": testMultiStage,
		"back":             backAll,
	}

	breakerRoutes = []string{
		"back",
	}

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
				msg := services.NewMessage(update.Message.Chat.ID, trans("miss_username", "en-US"))
				services.SendMessage(msg)
				continue
			}

			routing(update)
		}
	}
}
