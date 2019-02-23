package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

//Player struct
type Player struct {
	gorm.Model
	Username   string
	Inventory  Inventory
	State      []PlayerState
	Language   Language
	LanguageID uint
}

// Create player
func (p *Player) Create() *Player {
	services.Database.Create(&p)
	p.Inventory.Create()
	return p
}

// Update player
func (p *Player) Update() *Player {
	services.Database.Save(&p)
	p.Inventory.Update()
	return p
}

// Delete player
func (p *Player) Delete() *Player {
	services.Database.Delete(&p)
	p.Inventory.Delete()
	return p
}

//GetStateByFunction -
func (p *Player) GetStateByFunction(function string) PlayerState {
	var playerState PlayerState
	for _, state := range p.State {
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
