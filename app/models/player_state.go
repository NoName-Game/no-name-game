package models

import (
	"time"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// PlayerState -
type PlayerState struct {
	gorm.Model
	FinishAt  time.Time
	Function  string
	Payload   string
	PlayerID  uint
	Player    Player
	Stage     int
	ToNotify  bool
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

// GetAllStateToNotify - Get all rows from db to notify
func GetAllStateToNotify() (playerState []PlayerState) {
	services.Database.Where("completed = ?", false).Where("to_notify = ?", true).Where("finish_at < ?", time.Now()).Find(&playerState)

	return
}
