package models

import (
	"bitbucket.org/no-name-game/no-name/services"
	"github.com/jinzhu/gorm"
)

// Inventory -
type Inventory struct {
	gorm.Model
	Items []Item
}

func GetInvByPlayer(player Player) Inventory {
	var inventory Inventory
	services.Database.Set("gorm:auto_preload", true).Where("owner = ?", player).First(&inventory)

	return inventory
}

// Add an item
func (i *Inventory) addItem(item Item) *Inventory {
	i.Items = append(i.Items, item)

	return i
}

// Remove an item
func (i *Inventory) removeItem(item Item) *Inventory {
	for x := 0; x < len(i.Items); x++ {
		if i.Items[x] == item {
			i.Items[len(i.Items)-1], i.Items[x] = i.Items[x], i.Items[len(i.Items)-1]
			return i
		}
	}
	return i
}

// Create Inventory
func (i *Inventory) Create() *Inventory {
	services.Database.Create(&i)

	return i
}

// Update Inventory
func (i *Inventory) Update() *Inventory {
	services.Database.Save(&i)

	return i
}

// Delete Inventory
func (i *Inventory) Delete() *Inventory {
	services.Database.Delete(&i)

	return i
}
