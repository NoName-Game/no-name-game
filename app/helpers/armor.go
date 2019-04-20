package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/acme/namer"
	"bitbucket.org/no-name-game/no-name/app/models"
)

// NewCraftedArmor - Generate crafted armor
func NewCraftedArmor(c string) (armor models.Armor) {
	rarity := models.GetRarityBySlug("VC")
	category := models.GetArmorCategoryBySlug(c)

	armor = models.Armor{
		Name:          namer.GenerateName("resources/namer/armors/model.json"),
		Rarity:        rarity,
		ArmorCategory: category,
	}

	armor.Create()

	return
}
