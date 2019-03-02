package models

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Inventory - inventory struct
type Inventory struct {
	gorm.Model
	Items string //map[Item.ID]quantity
}

// AddItem - Add an item
func (i *Inventory) AddItem(item Item, quantity int) *Inventory {
	mapInventory := unmarshalInventory(i.Items)
	mapInventory[item.ID] += quantity

	i.Items = marshalInventory(mapInventory)
	i.Update()

	return i
}

// RemoveItem - Remove an item
func (i *Inventory) RemoveItem(item Item, quantity int) *Inventory {
	mapInventory := unmarshalInventory(i.Items)
	if _, ok := mapInventory[item.ID]; ok {
		mapInventory[item.ID] += -quantity // Decrement quantity
		if 0 >= mapInventory[item.ID] {    // Remove key if quantity is zero
			delete(mapInventory, item.ID)
		}
	}

	i.Items = marshalInventory(mapInventory)
	i.Update()

	return i
}

// Update - Update inventory
func (i *Inventory) Update() {
	services.Database.Save(&i)
}

// ToString - return inventory like a string
func (i *Inventory) ToString() string {
	var result string
	mapInventory := unmarshalInventory(i.Items)
	for key, value := range mapInventory {
		result += strconv.Itoa(value) + "x " + GetItemByID(key).Name + "\n"
	}

	return result
}

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

// marshalInventory - marshal player inventory
func marshalInventory(items map[uint]int) (inventory string) {
	inventoryJSON, err := json.Marshal(items)
	if err != nil {
		services.ErrorHandler("Error marshal inventory", err)
	}
	inventory = string(inventoryJSON)

	return
}
