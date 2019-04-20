package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

//Player struct
type Player struct {
	gorm.Model
	Username    string
	ChatID      int64
	States      []PlayerState
	Stars       []PlayerStar
	Positions   []PlayerPosition
	Ships       []PlayerShip
	Weapons     []Weapon
	Armors      []Armor
	Stats       PlayerStats
	StatsID     uint
	Language    Language
	LanguageID  uint
	Inventory   Inventory
	InventoryID uint
}

// Create player
func (p *Player) Create() *Player {
	services.Database.Create(&p)

	return p
}

// Update player
func (p *Player) Update() *Player {
	services.Database.Save(&p)

	return p
}

// Delete player
func (p *Player) Delete() *Player {
	services.Database.Delete(&p)

	return p
}

// GetStateByFunction -
func (p *Player) GetStateByFunction(function string) PlayerState {
	var playerState PlayerState
	services.Database.Where("function = ?", function).First(&playerState)

	return playerState
}

// FindPlayerByID - find player by ID
func FindPlayerByID(id uint) Player {
	var player Player
	services.Database.Where("id = ?", id).First(&player)

	return player
}

// GetEquippedArmors - get equipped player armors
func (p *Player) GetEquippedArmors() (armors Armors) {
	services.Database.Where("player_id = ? AND equipped = ?", p.ID, true).First(&armors)

	return
}

// GetEquippedWeapons - get equipped player weapons
func (p *Player) GetEquippedWeapons() (weapons Weapons) {
	services.Database.Where("player_id = ? AND equipped = ?", p.ID, true).First(&weapons)

	return
}
