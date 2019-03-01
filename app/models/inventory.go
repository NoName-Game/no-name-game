package models

import (
	"encoding/json"
)

// Inventory -
type Inventory struct {
	Items map[uint]int //map[Item.ID]quantity
}

// Add an item
func (i *Inventory) AddItem(item Item) *Inventory {
	i.Items[item.ID] += 1 //Aumento il valore

	return i
}

// Remove an item
func (i *Inventory) RemoveItem(item Item) *Inventory {
	if _, ok := i.Items[item.ID]; ok {
		i.Items[item.ID] -= 1      //Decremento la quantità
		if 0 == i.Items[item.ID] { // Se la quantità è pari a 0 elimino la key
			delete(i.Items, item.ID)
		}
	}
	return i
}

func (i *Inventory) Update(player Player) {
	savedInventory, _ := json.Marshal(i)
	player.Inventory = string(savedInventory)
	player.Update()
}

func (i *Inventory) ToString() string {
	var result string
	for key, value := range i.Items {
		result += string(value) + "x " + GetItemByID(key).Name + "\n"
	}
	return result
}
