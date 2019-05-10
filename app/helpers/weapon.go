package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/namer"
	"bitbucket.org/no-name-game/no-name/app/models"
)

// NewWeapon - Generate starter ship
func NewWeapon() (weapon models.Weapon) {
	rarity := models.GetRarityBySlug("VC")
	category := models.GetWeaponCategoryBySlug("knife")

	var name string
	switch category.Slug {
	case "knife":
		name = namer.GenerateName("resources/namer/weapons/knife/model.json")
	default:
		name = namer.GenerateName("resources/namer/weapons/model.json")
	}

	weapon = models.Weapon{
		Name:           name,
		Rarity:         rarity,
		WeaponCategory: category,
	}

	weapon.Create()

	return
}

// NewCraftedWeapon - Generate starter ship
func NewCraftedWeapon(c string) (weapon models.Weapon) {
	rarity := models.GetRarityBySlug("VC")
	category := models.GetWeaponCategoryBySlug(c)

	var name string
	switch category.Slug {
	case "knife":
		name = namer.GenerateName("resources/namer/weapons/knife/model.json")
	default:
		name = namer.GenerateName("resources/namer/weapons/model.json")
	}

	weapon = models.Weapon{
		Name:           name,
		Rarity:         rarity,
		WeaponCategory: category,
	}

	weapon.Create()

	return
}
