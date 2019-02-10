package app

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// PlayerState -
type PlayerState struct {
	gorm.Model
	PlayerID uint
	Function string
	Stage    int
	Payload  string
}

// Create Player State
func (s *PlayerState) create() *PlayerState {
	services.Database.Create(&s)

	return s
}

// Create Player State
func (s *PlayerState) update() *PlayerState {
	services.Database.Save(&s)

	return s
}
