package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// PlayerShip
type PlayerShip struct {
	gorm.Model
	PlayerID uint
	Ship     Ship
	ShipID   uint
}

// Create PlayerShip
func (ps *PlayerShip) Create() *PlayerShip {
	services.Database.Create(&ps)

	return ps
}

// Update PlayerShip
func (ps *PlayerShip) Update() *PlayerShip {
	services.Database.Save(&ps)

	return ps
}

// Delete Player state
func (ps *PlayerShip) Delete() *PlayerShip {
	services.Database.Delete(&ps)

	return ps
}
