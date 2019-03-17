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

// FindPlayerByUsername - find player by username
func FindPlayerByUsername(username string) Player {
	var player Player
	services.Database.Where("username = ?", username).First(&player)

	return player
}

// FindPlayerByID - find player by ID
func FindPlayerByID(id uint) Player {
	var player Player
	services.Database.Where("id = ?", id).First(&player)

	return player
}

// AddStar - Associate star to player
func (p *Player) AddStar(star Star) *Player {
	services.Database.Model(&p).Association("Stars").Append(PlayerStar{Star: star})

	return p
}

// AddPosition
func (p *Player) AddPosition(position PlayerPosition) *Player {
	services.Database.Model(&p).Association("Positions").Append(position)

	return p
}

// AddShip - Associate ship to player
func (p *Player) AddShip(ship Ship) *Player {
	services.Database.Model(&p).Association("Ships").Append(PlayerShip{Ship: ship})

	return p
}
