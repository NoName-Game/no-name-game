package models

import (
	"encoding/json"
)

// Inventory -
type Inventory struct {
	Items []Item
}

// Add an item
func (i *Inventory) AddItem(item Item) *Inventory {
	i.Items = append(i.Items, item)

	return i
}

// Remove an item
func (i *Inventory) RemoveItem(item Item) *Inventory {
	for x := 0; x < len(i.Items); x++ {
		if item == i.Items[x] {
			i.Items[x] = i.Items[len(i.Items)-1]
			i.Items[len(i.Items)-1] = Item{Name: ""} // Elimino il valore.
			i.Items = i.Items[:len(i.Items)-1]       // Tronco l'array.
			return i
		}
	}
	return i
}

func (i *Inventory) Update(player Player) {
	savedInventory, _ := json.Marshal(i)
	player.Inventory = string(savedInventory)
	player.Update()
}
