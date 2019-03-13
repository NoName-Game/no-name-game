package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/namer"
	"bitbucket.org/no-name-game/no-name/app/models"
)

// NewShip - Generate new ship
func NewShip() (ship models.Ship) {
	ship = models.Ship{
		Name: namer.GenerateName("resources/stars/model.json"),
	}
	//FIXME: todo

	ship.Create()

	return
}
