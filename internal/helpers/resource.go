package helpers

import (
	"sort"

	"nn-grpc/build/pb"
)

// GetResourceCategoryIcons
func GetResourceCategoryIcons(categoryID uint32) (category string) {
	switch categoryID {
	case 1:
		category = "🔥"
	case 2:
		category = "💧"
	case 3:
		category = "⚡️"
	}
	return
}

// GetResourceBaseIcons
func GetResourceBaseIcons(isBase bool) (result string) {
	if isBase {
		result = "🔬Base"
	}
	return
}

// SortInventoryByRarity
func SortInventoryByRarity(inventory []*pb.PlayerInventory) []*pb.PlayerInventory {
	sort.Slice(inventory, func(i, j int) bool {
		return inventory[i].GetResource().GetRarityID() < inventory[j].GetResource().GetRarityID()
	})

	return inventory
}
