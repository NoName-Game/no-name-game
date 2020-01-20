package controllers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

// Writer: reloonfire
// Starting on: 17/01/2020
// Project: no-name-game

// Crafting:
// Craft Base effettuati dal player

//====================================
// CraftingController
//====================================
type CraftingV2Controller struct {
	BaseController
	Payload struct {
		Item      nnsdk.Crafted // Item da craftare
		Resources map[uint]int  // Materiali necessari
	}
}

//====================================
// Handle
//====================================
func (c *CraftingV2Controller) Handle(update tgbotapi.Update) {
	// Current Controller instance
	var err error
	var isNewState bool
	c.RouteName, c.Update, c.Message = "route.crafting", update, update.Message

	// Check current state for this routes
	c.State, isNewState = helpers.CheckState(c.RouteName, c.Payload, helpers.Player)

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// It's first message
	if isNewState {
		c.Stage()
		return
	}

	// Go to validator
	if !c.Validator() {
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}

		// Ok! Run!
		c.Stage()
		return
	}

	// Validator goes errors
	validatorMsg := services.NewMessage(c.Message.Chat.ID, c.Validation.Message)
	services.SendMessage(validatorMsg)
	return
}

//====================================
// Validator
//====================================
func (c *CraftingV2Controller) Validator() (hasErrors bool) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch c.State.Stage {
	case 0:
		if strings.Contains(c.Message.Text, helpers.Trans("crafting.craft")) {
			c.State.Stage = 1
			return false
		}
	}

	return true
}

//====================================
// Stage  0 What -> 1 - Check Resources -> 2 - Confirm -> 3 - Craft
//====================================
func (c *CraftingV2Controller) Stage() {
	var err error

	switch c.State.Stage {
	case 0:
		// Lista oggetti craftabili
		// TODO: sulla base di ciò che ha espresso il gruppo, il player può craftare solo item di rarità <R

		craftableItems, err := providers.GetAllCraftableItems()

		if err != nil {
			services.ErrorHandler("Can't retrieve CraftableItems", err)
		}

		msg := helpers.ListItems(craftableItems)

		services.SendMessage(msg)
	case 1:
		// Check Resources
		c.Payload.Item, err = providers.GetCraftedByName(strings.Split(c.Message.Text, ": ")[1])

		if err != nil {
			services.ErrorHandler("Can't retrieve CraftableItems by Name", err)
		}

		helpers.UnmarshalPayload(c.Payload.Item.Recipe.RecipeList, &c.Payload.Resources)

		msg := services.NewMessage(helpers.Player.ChatID, helpers.Trans("crafting.you_need", c.Payload.Item.Item.Name, helpers.ListRecipe(c.Payload.Resources)))

	}
}
