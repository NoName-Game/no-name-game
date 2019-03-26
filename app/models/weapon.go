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
	PlayerID         uint // Has Many from user
	Equipped         bool
}

// Weapons - Weapons struct
type Weapons []Weapon

// Create Weapon
func (w *Weapon) Create() *Weapon {
	services.Database.Create(&w)

	return w
}

// Update Weapon
func (w *Weapon) Update() *Weapon {
	services.Database.Save(&w)

	return w
}

// Delete Weapon
func (w *Weapon) Delete() *Weapon {
	services.Database.Delete(&w)

	return w
}

// AddPlayer
func (w *Weapon) AddPlayer(player Player) *Weapon {
	w.PlayerID = player.ID
	services.Database.Save(&w)

	return w
}

// GetWeaponByName - Get Weapon by name
func GetWeaponByName(name string) Weapon {
	var weapon Weapon
	services.Database.Set("gorm:auto_preload", true).Where("name = ?", name).First(&weapon)

	return weapon
}

// GetWeaponByID - Get weapon by ID
func GetWeaponByID(id uint) Weapon {
	var weapon Weapon
	services.Database.Set("gorm:auto_preload", true).Where("id = ?", id).First(&weapon)

	return weapon
}
