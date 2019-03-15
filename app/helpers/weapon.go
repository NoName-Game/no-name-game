package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/namer"
	"bitbucket.org/no-name-game/no-name/app/models"
)

// NewWeapon - Generate starter ship
func NewWeapon() (weapon models.Weapon) {
	rarity := models.GetRarityBySlug("VC")
	category := models.GetWeaponCategoryBySlug("knife")
	weapon = models.Weapon{
		Name:           namer.GenerateName("resources/namer/weapons/model.json"),
		Rarity:         rarity,
		WeaponCategory: category,
	}

	weapon.Create()

	return
}
