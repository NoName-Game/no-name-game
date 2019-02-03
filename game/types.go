package game

import (
	"github.com/jinzhu/gorm"
	"gitlab.com/Valkyrie00/no-name/config"
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

// PlayerState -
type PlayerState struct {
	gorm.Model
	PlayerID int
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
