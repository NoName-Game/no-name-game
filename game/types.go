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
	State    PlayerState
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

// FindByUsername - find player by username
func findPlayerByUsername(player Player, username string) {
	config.Database.Where("username = ?", username).First(&player)
}

// PlayerState -
type PlayerState struct {
	gorm.Model
	PlayerID uint
	Function string
	Stage    int
	Payload  string
}

// Create player
func (s *PlayerState) create() *PlayerState {
	config.Database.Create(s)

	return s
}

func (s *PlayerState) update() *PlayerState {
	config.Database.Save(s)

	return s
}
