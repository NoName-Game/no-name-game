package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Ship - Ship struct
type Ship struct {
	gorm.Model
	Name           string `json:"name"`
	Rarity         Rarity
	RarityID       uint
	ShipCategory   ShipCategory
	ShipCategoryID uint
}

// Ships - Ships struct
type Ships []Ship

// Create ship
func (s *Ship) Create() *Ship {
	services.Database.Create(&s)

	return s
}
