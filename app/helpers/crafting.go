package helpers

import (
	"bitbucket.org/no-name-game/no-name/app/models"
)

// Crafting - this method generate items by craft
func Crafting(maplist string) (craftingResult string) {
	type craftingPayload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}

	var payload craftingPayload
	UnmarshalPayload(maplist, &payload)

	// var craftingResult string
	switch payload.Item {
	case "armors":
		var armor models.Armor
		armor = NewCraftedArmor(Slugger(payload.Category))
		craftingResult = armor.Name
	case "weapons":
		var weapon models.Weapon
		weapon = NewCraftedWeapon(Slugger(payload.Category))
		craftingResult = weapon.Name
	}

	//TODO: associate armor or weapon to player

	return
}
