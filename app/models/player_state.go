package models

import (
	"time"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// PlayerState -
type PlayerState struct {
	gorm.Model
	PlayerID  uint
	Function  string
	Stage     int
	FinishAt  time.Time
	ToNotify  bool
	Payload   string
	Completed bool `gorm:"default: false"`
}

// Create Player State
func (s *PlayerState) Create() *PlayerState {
	services.Database.Create(&s)

	return s
}

// Update Player State
func (s *PlayerState) Update() *PlayerState {
	services.Database.Save(&s)

	return s
}

// Delete Player state
func (s *PlayerState) Delete() *PlayerState {
	services.Database.Delete(&s)

	return s
}
