package app

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

//******************************
// Player
//******************************

//Player struct
type Player struct {
	gorm.Model
	Username   string
	State      []PlayerState
	Language   Language
	LanguageID uint
}

// Create player
func (p *Player) create() *Player {
	services.Database.Create(&p)

	return p
}

// Update player
func (p *Player) update() *Player {
	services.Database.Save(&p)

	return p
}

// Delete player
func (p *Player) delete() *Player {
	services.Database.Delete(&p)

	return p
}

func (p *Player) getStateByFunction(function string) PlayerState {
	var playerState PlayerState
	for _, state := range p.State {
		if state.Function == function {
			return state
		}
	}

	return playerState
}

// FindByUsername - find player by username
func findPlayerByUsername(username string) Player {
	var player Player
	services.Database.Set("gorm:auto_preload", true).Where("username = ?", username).First(&player)

	return player
}

// PlayerState -
type PlayerState struct {
	gorm.Model
	PlayerID  uint
	Function  string
	Stage     int
	Payload   string
	Completed bool `gorm:"default: false"`
}

// Create Player State
func (s *PlayerState) create() *PlayerState {
	services.Database.Create(&s)

	return s
}

// Create Player State
func (s *PlayerState) update() *PlayerState {
	services.Database.Save(&s)

	return s
}

func (s *PlayerState) delete() *PlayerState {
	services.Database.Delete(&s)

	return s
}

// Language -
type Language struct {
	gorm.Model
	Language string
}

// getDefaultLangID - get Default Lang ID
func getDefaultLangID(lang string) Language {
	var language Language
	services.Database.Set("gorm:auto_preload", true).Where("language = ?", lang).First(&language)

	return language
}

func seederLanguage() {
	for _, lang := range services.Langs {
		newLanguage := Language{Language: lang}
		services.Database.Where(Language{Language: lang}).FirstOrCreate(&newLanguage)
	}
}
