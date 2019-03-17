package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Armor - Armor struct
type Armor struct {
	gorm.Model
	Name            string `json:"name"`
	Rarity          Rarity
	RarityID        uint
	ArmorCategory   ArmorCategory
	ArmorCategoryID uint
	PlayerID        uint
}

// Armors - Armors struct
type Armors []Armor

// Create Armor
func (a *Armor) Create() *Armor {
	services.Database.Create(&a)

	return a
}

// Update Armor
func (a *Armor) Update() *Armor {
	services.Database.Save(&a)

	return a
}

// AddPlayer
func (a *Armor) AddPlayer(player Player) *Armor {
	a.PlayerID = player.ID
	services.Database.Save(&a)

	return a
}
