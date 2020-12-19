package helpers

import (
	"log"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
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
	var n = len(inventory)
	for i := 1; i < n; i++ {
		j := i
		for j > 0 {
			if inventory[j-1].GetResource().GetRarityID() > inventory[j].GetResource().GetRarityID() {
				inventory[j-1], inventory[j] = inventory[j], inventory[j-1]
			}
			j = j - 1
		}
	}

	log.Println(inventory)
	return inventory
}
