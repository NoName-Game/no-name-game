package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// PlayerPosition -
type PlayerPosition struct {
	gorm.Model
	PlayerID uint
	X        float64
	Y        float64
	Z        float64
}

// Create PlayerPosition
func (p *PlayerPosition) Create() *PlayerPosition {
	services.Database.Create(&p)

	return p
}

// Update PlayerPosition
func (p *PlayerPosition) Update() *PlayerPosition {
	services.Database.Save(&p)

	return p
}

// Delete Player state
func (p *PlayerPosition) Delete() *PlayerPosition {
	services.Database.Delete(&p)

	return p
}
