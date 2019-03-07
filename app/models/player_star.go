package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// PlayerStar
type PlayerStar struct {
	gorm.Model
	PlayerID uint
	Star     Star
	StarID   uint
}

// Create PlayerStar
func (ps *PlayerStar) Create() *PlayerStar {
	services.Database.Create(&ps)

	return ps
}

// Update PlayerStar
func (ps *PlayerStar) Update() *PlayerStar {
	services.Database.Save(&ps)

	return ps
}

// Delete Player state
func (ps *PlayerStar) Delete() *PlayerStar {
	services.Database.Delete(&ps)

	return ps
}
