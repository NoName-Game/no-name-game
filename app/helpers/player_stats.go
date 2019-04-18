package helpers

import (
	"log"

	"bitbucket.org/no-name-game/no-name/app/models"
	"bitbucket.org/no-name-game/no-name/services"
)

// DecrementLife - Handle the life points
func DecrementLife(lifePoint uint, player models.Player) {

	// MaxLife = 100 + Level * 10

	if player.Stats.LifePoint-lifePoint > 100+player.Stats.Level*10 { // Overflow problem
		player.Stats.LifePoint = 0
	} else {
		player.Stats.LifePoint -= lifePoint
	}
	player.Stats.Update()
	if player.Stats.LifePoint == 0 {
		// Player Die
		log.Println(player.States)
		for _, state := range player.States {
			log.Println(state.Function)
			FinishAndCompleteState(state, player)
		}
		msg := services.NewMessage(player.ChatID, Trans("playerDie", player.Language.Slug))
		msg.ParseMode = "HTML"
		services.SendMessage(msg)
	}
}
