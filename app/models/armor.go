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
	Equipped        bool
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

// Delete Armor
func (a *Armor) Delete() *Armor {
	services.Database.Delete(&a)

	return a
}

// AddPlayer
func (a *Armor) AddPlayer(player Player) *Armor {
	a.PlayerID = player.ID
	services.Database.Save(&a)

	return a
}

// GetArmorByName - Get Armor by name
func GetArmorByName(name string) Armor {
	var armor Armor
	services.Database.Set("gorm:auto_preload", true).Where("name = ?", name).First(&armor)

	return armor
}

// GetArmorByID - Get armor by ID
func GetArmorByID(id uint) Armor {
	var armor Armor
	services.Database.Set("gorm:auto_preload", true).Where("id = ?", id).First(&armor)

	return armor
}
