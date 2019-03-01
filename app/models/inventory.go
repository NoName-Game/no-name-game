package models

import (
	"encoding/json"
)

// Inventory - inventory struct
type Inventory struct {
	Items map[uint]int //map[Item.ID]quantity
}

// AddItem - Add an item
func (i *Inventory) AddItem(item Item) *Inventory {
	i.Items[item.ID]++ // Increment quantity

	return i
}

// RemoveItem - Remove an item
func (i *Inventory) RemoveItem(item Item) *Inventory {
	if _, ok := i.Items[item.ID]; ok {
		i.Items[item.ID]--         // Decrement quantity
		if 0 == i.Items[item.ID] { // Remove key if quantity is zero
			delete(i.Items, item.ID)
		}
	}

	return i
}

// Update - Update inventory
func (i *Inventory) Update(player Player) {
	savedInventory, _ := json.Marshal(i)
	player.Inventory = string(savedInventory)
	player.Update()
}

// ToString - return inventory like a string
func (i *Inventory) ToString() string {
	var result string
	for key, value := range i.Items {
		result += string(value) + "x " + GetItemByID(key).Name + "\n"
	}

	return result
}
