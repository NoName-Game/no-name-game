package helpers

import "bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

func InventoryResourcesToMap(inventory nnsdk.PlayerInventories) map[uint]int {
	inventoryMap := make(map[uint]int)
	for i := 0; i < len(inventory); i++ {
		if inventory[i].ItemType == "resources" {
			inventoryMap[inventory[i].ItemID] = *inventory[i].Quantity
		}
	}
	return inventoryMap
}
