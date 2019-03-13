package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/namer"
	"bitbucket.org/no-name-game/no-name/app/models"
)

// NewStartShip - Generate starter ship
func NewStartShip() (ship models.Ship) {
	rarity := models.GetRarityBySlug("VC")
	category := models.GetShipCategoryBySlug("normal")
	ship = models.Ship{
		Name:         namer.GenerateName("resources/namer/ships/model.json"),
		Rarity:       rarity,
		ShipCategory: category,
	}

	ship.Create()

	return
}
