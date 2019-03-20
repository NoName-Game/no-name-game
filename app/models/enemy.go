package models

import (
	"github.com/jinzhu/gorm"
)

// Enemy - Rappresenta un nemico generico
type Enemy struct {
	gorm.Model
	Helmet           Armor
	Chestplate       Armor
	Boots            Armor
	Name             string
	Life             int32
	DamageMultiplier float32
}
