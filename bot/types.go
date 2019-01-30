package bot

import "github.com/jinzhu/gorm"

//******************************
// User
//******************************

//User - user struct
type User struct {
	gorm.Model
	Username string
}
