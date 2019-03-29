package models

import (
	"fmt"
	"reflect"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

/*Le caratteristiche del giocatore sono:
FOR: Forza
DES: Destrezza
COS: Costituzione
INT: Intelligenza
SAG: Saggezza
CAR: Carisma*/

type PlayerStats struct {
	gorm.Model
	Forza        uint `gorm:"default:1"`
	Destrezza    uint `gorm:"default:1"`
	Costituzione uint `gorm:"default:1"`
	Intelligenza uint `gorm:"default:1"`
	Saggezza     uint `gorm:"default:1"`
	Carisma      uint `gorm:"default:1"`
}

// Create Player State
func (s *PlayerStats) Create() *PlayerStats {
	services.Database.Create(&s)

	return s
}

// Update Player State
func (s *PlayerStats) Update() *PlayerStats {
	services.Database.Save(&s)

	return s
}

// Delete Player state
func (s *PlayerStats) Delete() *PlayerStats {
	services.Database.Delete(&s)

	return s
}

func (s *PlayerStats) ToString() (result string) {
	val := reflect.ValueOf(s).Elem()
	result += "Ecco le tue statistiche:\n"
	for i := 1; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		result += fmt.Sprintf("<code>%-15v:%v</code>\n", typeField.Name, valueField.Interface())
	}
	return
}
