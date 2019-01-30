package game

import "github.com/jinzhu/gorm"

//******************************
// Player
//******************************

//Player struct
type Player struct {
	gorm.Model
	Username string
}
