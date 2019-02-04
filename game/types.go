package game

import (
	"bitbucket.org/no-name-game/no-name/config"
	"github.com/jinzhu/gorm"
)

//******************************
// Player
//******************************

//Player struct
type Player struct {
	gorm.Model
	Username string
}

// Create player
func (p *Player) create() *Player {
	config.Database.Create(p)

	return p
}

// Update player
func (p *Player) update() *Player {
	config.Database.Save(p)

	return p
}

// Delete player
func (p *Player) delete() *Player {
	config.Database.Delete(p)

	return p
}

// FindByUsername -find player by username
func (p *Player) findByUsername(username string) *Player {
	config.Database.Where("username = ?", username).First(p)

	return p
}
