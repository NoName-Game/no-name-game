package controllers

import (
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

//====================================
// Inventory
//====================================
type InventoryController BaseController

//====================================
// Handle
//====================================
func (c *InventoryController) Handle(update tgbotapi.Update) {
	c.Message = update.Message

	msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.intro"))
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.inventory.recap")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.inventory.equip")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.inventory.destroy")),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		),
	)

	services.SendMessage(msg)
	return
}

//====================================
// Inventory Recap
//====================================

type InventoryRecapController BaseController

//====================================
// Handle
//====================================
func (c *InventoryRecapController) Handle(update tgbotapi.Update) {
	c.Message = update.Message

	var recap string

	// Summary Resources
	playerInventory, err := providers.GetPlayerInventory(helpers.Player)
	if err != nil {
		services.ErrorHandler("Can't get player inventory", err)
	}

	recap += "\n" + helpers.Trans("resources") + ":\n"
	playerResources := helpers.InventoryToMap(playerInventory)
	for r, q := range playerResources {
		resource, errResouce := providers.GetResourceByID(r)
		if errResouce != nil {
			services.ErrorHandler("Error in InventoryToString", err)
		}

		recap += "- " + resource.Name + " (" + (strconv.Itoa(q)) + ")\n"
	}

	// Summary Weapons
	playerWeapons, errWeapons := providers.GetPlayerWeapons(helpers.Player, "false")
	if errWeapons != nil {
		services.ErrorHandler("Can't get player weapons", err)
	}
	recap += "\n" + helpers.Trans("weapons") + ":\n"
	for _, weapon := range playerWeapons {
		recap += "- " + weapon.Name + "\n"
	}

	// Summary Armors
	playerArmors, errArmors := providers.GetPlayerArmors(helpers.Player, "false")
	if errArmors != nil {
		services.ErrorHandler("Can't get player armors", err)
	}

	recap += "\n" + helpers.Trans("armors") + ":\n"
	for _, armor := range playerArmors {
		recap += "- " + armor.Name + "\n"
	}

	msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("inventory.recap")+recap)
	services.SendMessage(msg)
	return
}
