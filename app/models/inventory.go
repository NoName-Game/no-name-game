package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Inventory - inventory struct
type Inventory struct {
	gorm.Model
	Items string //map[Item.ID]quantity
}

// Create inventory
func (i *Inventory) Create() *Inventory {
	services.Database.Create(&i)

	return i
}

// Update - Update inventory
func (i *Inventory) Update() {
	services.Database.Save(&i)
}
