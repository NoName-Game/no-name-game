package app

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

//Player struct
type Player struct {
	gorm.Model
	Username string
	State    PlayerState
}

// Create player
func (p *Player) create() *Player {
	services.Database.Create(&p)

	return p
}

// Update player
func (p *Player) update() *Player {
	services.Database.Save(&p)

	return p
}

// Delete player
func (p *Player) delete() *Player {
	services.Database.Delete(&p)

	return p
}

// FindByUsername - find player by username
func findPlayerByUsername(username string) Player {
	var player Player
	services.Database.Set("gorm:auto_preload", true).Where("username = ?", username).First(&player)

	return player
}
