package helpers

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

// unmarshalInventory - Unmarshal player inventory
func unmarshalInventory(inventory string) (items map[uint]int) {
	if inventory != "" {
		err := json.Unmarshal([]byte(inventory), &items)
		if err != nil {
			services.ErrorHandler("Error unmarshal inventory", err)
		}
	} else {
		items = make(map[uint]int)
	}

	return
}

// ToString - return inventory like a string
func InventoryToString(i nnsdk.Inventory) string {
	var result string
	mapInventory := unmarshalInventory(i.Items)
	for key, value := range mapInventory {
		resource, err := providers.GetResourceByID(key)
		if err != nil {
			services.ErrorHandler("Error in InventoryToString", err)
		}

		result += strconv.Itoa(value) + "x " + resource.Name + "\n"
	}

	return result
}

// ToMap - return inventory unmarshal
func InventoryToMap(i nnsdk.Inventory) (items map[uint]int) {
	items = unmarshalInventory(i.Items)

	return
}

// ToKeyboardAddCraft - return inventory for keyboard
func InventoryToKeyboardAddCraft(i nnsdk.Inventory) (results []string) {
	mapInventory := unmarshalInventory(i.Items)
	for key, value := range mapInventory {
		resource, err := providers.GetResourceByID(key)
		if err != nil {
			services.ErrorHandler("Error in ToKeyBoardAddCraft", err)
		}

		results = append(results, "Add "+resource.Name+" ("+strconv.Itoa(value)+")")
	}

	return
}
