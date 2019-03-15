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
}

// Armors - Armors struct
type Armors []Armor

// Create Armor
func (w *Armor) Create() *Armor {
	services.Database.Create(&w)

	return w
}
