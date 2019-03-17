package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/models"
)

// CraftArmor - this method generate armor for crafting controller
func CraftArmor(maplist string) (armor models.Armor) {
	type craftingPayload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}

	var payload craftingPayload
	UnmarshalPayload(maplist, &payload)

	armor = NewCraftedArmor(Slugger(payload.Category))

	return
}

// CraftWeapon - this method generate weapon for crafting controller
func CraftWeapon(maplist string) (weapon models.Weapon) {
	type craftingPayload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}

	var payload craftingPayload
	UnmarshalPayload(maplist, &payload)

	weapon = NewCraftedWeapon(Slugger(payload.Category))

	return
}
