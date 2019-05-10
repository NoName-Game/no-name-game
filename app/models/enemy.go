package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Enemy - Enemy struct
type Enemy struct {
	gorm.Model
	Name      string `json:"name"`
	LifePoint uint
}

// Enemies - Enemies struct
type Enemies []Enemy

// Create Enemy
func (e *Enemy) Create() *Enemy {
	services.Database.Create(&e)

	return e
}

// Update Armor
func (e *Enemy) Update() *Enemy {
	services.Database.Save(&e)

	return e
}

// Delete Armor
func (e *Enemy) Delete() *Enemy {
	services.Database.Delete(&e)

	return e
}

// GetEnemyByID - Get enemy by ID
func GetEnemyByID(id uint) Enemy {
	var enemy Enemy
	services.Database.Set("gorm:auto_preload", true).Where("id = ?", id).First(&enemy)

	return enemy
}
