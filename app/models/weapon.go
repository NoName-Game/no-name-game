package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Weapon - Weapon struct
type Weapon struct {
	gorm.Model
	Name             string `json:"name"`
	Rarity           Rarity
	RarityID         uint
	WeaponCategory   WeaponCategory
	WeaponCategoryID uint
}

// Weapons - Weapons struct
type Weapons []Weapon

// Create Weapon
func (w *Weapon) Create() *Weapon {
	services.Database.Create(&w)

	return w
}
