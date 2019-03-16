package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

//Player struct
type Player struct {
	gorm.Model
	Username    string
	States      []PlayerState
	Stars       []PlayerStar
	Positions   []PlayerPosition
	Ships       []PlayerShip
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
	for _, state := range p.States {
		if state.Function == function {
			return state
		}
	}

	return playerState
}

// FindPlayerByUsername - find player by username
func FindPlayerByUsername(username string) Player {
	var player Player
	services.Database.Set("gorm:auto_preload", true).Where("username = ?", username).First(&player)

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
