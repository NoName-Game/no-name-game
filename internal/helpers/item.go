package helpers

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
)

// SortInventoryByRarity
func SortItemByCategory(inventory []*pb.PlayerInventory) []*pb.PlayerInventory {
	sort.Slice(inventory, func(i, j int) bool {
		return inventory[i].GetItem().GetItemCategoryID() < inventory[j].GetItem().GetItemCategoryID()
	})

	return inventory
}

func WeaponFormatter(weapon *pb.Weapon) string {
	return fmt.Sprintf(
		"<b>%s</b> (%s) - [%v, %v%%, %v] 🎖%v",
		weapon.Name,
		strings.ToUpper(weapon.Rarity.Slug),
		math.Round(weapon.RawDamage),
		math.Round(weapon.Precision),
		weapon.Durability,
		weapon.Rarity.LevelToEuip,
	)
}

func ArmorFormatter(armor *pb.Armor) string {
	return fmt.Sprintf(
		"\n<b>%s</b> (%s) - [%v, %v%%, %v%%] 🎖%v",
		armor.Name,
		strings.ToUpper(armor.Rarity.Slug),
		math.Round(armor.Defense),
		math.Round(armor.Evasion),
		math.Round(armor.Halving),
		armor.Rarity.LevelToEuip,
	)
}
