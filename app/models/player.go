package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

//Player struct
type Player struct {
	gorm.Model
	Username   string
	State      []PlayerState
	Language   Language
	LanguageID uint
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
