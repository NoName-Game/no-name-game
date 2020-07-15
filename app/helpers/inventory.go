package helpers

import "bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"

func InventoryResourcesToMap(inventory nnsdk.PlayerInventories) map[uint]int {
	inventoryMap := make(map[uint]int)
	var item nnsdk.PlayerInventory
	for item = range inventory {
		if item.ItemType == "resources" {
			inventoryMap[item.ItemID] = *item.Quantity
		}
	}
	return inventoryMap
}
