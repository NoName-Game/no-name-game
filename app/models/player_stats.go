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
	Esperienza   uint `gorm:"default:0"`
	Livello      uint `gorm:"default:1"`
	Forza        uint `gorm:"default:1"`
	Destrezza    uint `gorm:"default:1"`
	Costituzione uint `gorm:"default:1"`
	Intelligenza uint `gorm:"default:1"`
	Saggezza     uint `gorm:"default:1"`
	Carisma      uint `gorm:"default:1"`
	AbilityPoint uint `gorm:"default:10"`
}

// Create Player State
func (s *PlayerStats) Create() *PlayerStats {
	services.Database.Create(&s)

	return s
}

// Update Player State
func (s *PlayerStats) Update() *PlayerStats {
	if s.Esperienza >= 100 { //Controllo che l'esperienza posseduta sia abbastanza per aumentare di livello e assegno gli ability point
		s.Esperienza -= 100
		s.Livello++
		s.AbilityPoint++
	}
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
	for i := 3; i < val.NumField()-1; i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		result += fmt.Sprintf("<code>%-15v:%v</code>\n", typeField.Name, valueField.Interface())
	}
	return
}

func (s *PlayerStats) Increment(fieldName string) {
	f := reflect.ValueOf(s).Elem().FieldByName(fieldName)
	f.SetUint(uint64(f.Interface().(uint) + 1))
	s.AbilityPoint--
}
