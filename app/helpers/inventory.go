package helpers

import (
	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
)

func InventoryResourcesToMap(inventory []*pb.PlayerInventory) map[uint32]int32 {
	inventoryMap := make(map[uint32]int32)
	for i := 0; i < len(inventory); i++ {
		if inventory[i].ItemType == "resources" {
			inventoryMap[inventory[i].ItemID] = inventory[i].Quantity
		}
	}
	return inventoryMap
}
