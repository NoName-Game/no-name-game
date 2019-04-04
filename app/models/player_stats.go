package models

import (
	"fmt"
	"reflect"
	"strings"

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
	Experience   uint `gorm:"default:0"`
	Level        uint `gorm:"default:1"`
	Strength     uint `gorm:"default:1"`
	Dexterity    uint `gorm:"default:1"`
	Constitution uint `gorm:"default:1"`
	Intelligence uint `gorm:"default:1"`
	Wisdom       uint `gorm:"default:1"`
	Charisma     uint `gorm:"default:1"`
	AbilityPoint uint `gorm:"default:0"`
}

// Create Player State
func (s *PlayerStats) Create() *PlayerStats {
	services.Database.Create(&s)

	return s
}

// Update Player State
func (s *PlayerStats) Update() *PlayerStats {
	if s.Experience >= 100 { //Controllo che l'esperienza posseduta sia abbastanza per aumentare di livello e assegno gli ability point
		s.Experience -= 100
		s.Level++
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

func (s *PlayerStats) ToString(slug string) (result string) {
	val := reflect.ValueOf(s).Elem()
	for i := 3; i < val.NumField()-1; i++ {
		valueField := val.Field(i)
		fieldName, _ := services.GetTranslation("ability."+strings.ToLower(val.Type().Field(i).Name), slug, nil)

		result += fmt.Sprintf("<code>%-15v:%v</code>\n", fieldName, valueField.Interface())
	}
	return
}

func (s *PlayerStats) Increment(fieldName string) {
	f := reflect.ValueOf(s).Elem().FieldByName(fieldName)
	f.SetUint(uint64(f.Interface().(uint) + 1))
	s.AbilityPoint--
}
