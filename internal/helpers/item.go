package helpers

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"nn-grpc/build/pb"
)

// SortInventoryByRarity
func SortItemByCategory(inventory []*pb.PlayerInventory) []*pb.PlayerInventory {
	sort.Slice(inventory, func(i, j int) bool {
		return inventory[i].GetItem().GetItemCategoryID() < inventory[j].GetItem().GetItemCategoryID()
	})

	return inventory
}

func GetWeaponEfficencyBySlug(efficencySlug string) string {
	switch efficencySlug {
	case "fire":
		return "🔥"
	case "water":
		return "💧"
	case "electric":
		return "⚡️"
	case "void":
		return "🌀"
	}

	return ""
}

func WeaponFormatter(weapon *pb.Weapon) string {
	return fmt.Sprintf(
		"%s (%s) %s [%v, %v%%, %v] 🎖%v",
		weapon.Name,
		strings.ToUpper(weapon.Rarity.Slug),
		GetWeaponEfficencyBySlug(weapon.Efficency.Slug),
		math.Round(weapon.RawDamage),
		math.Round(weapon.Precision),
		weapon.Durability,
		weapon.Rarity.LevelToEuip,
	)
}

func ArmorFormatter(armor *pb.Armor) string {
	return fmt.Sprintf(
		"%s (%s) 🛡 [%v, %v%%, %v%%] 🎖%v",
		armor.Name,
		strings.ToUpper(armor.Rarity.Slug),
		math.Round(armor.Defense),
		math.Round(armor.Evasion),
		math.Round(armor.Halving),
		armor.Rarity.LevelToEuip,
	)
}
