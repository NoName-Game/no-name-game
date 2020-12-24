package helpers

import (
	"sort"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
)

// SortInventoryByRarity
func SortItemByCategory(inventory []*pb.PlayerInventory) []*pb.PlayerInventory {
	sort.Slice(inventory, func(i, j int) bool {
		return inventory[i].GetItem().GetItemCategoryID() < inventory[j].GetItem().GetItemCategoryID()
	})

	return inventory
}
