package controllers

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Crafting:
// I craft possono essere affidati a NPC o essere eseguiti dal player (solo craft semplici).

//====================================
// CraftingController
//====================================
type CraftingController struct {
	BaseController
	Payload struct {
		Item      string
		Category  string
		Resources map[uint]int
	}
	// Additional Data
	AddResourceFlag bool
}

//====================================
// Handle
//====================================
func (c *CraftingController) Handle(update tgbotapi.Update) {
	// Current Controller instance
	var err error
	var isNewState bool
	c.RouteName, c.Update, c.Message = "route.crafting", update, update.Message

	// Set Additional Data
	c.AddResourceFlag = false

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
func (c *CraftingController) Validator() (hasErrors bool) {
	c.Validation.Message = helpers.Trans("validationMessage")

	switch c.State.Stage {
	case 0:
		// Verifico se ha scelto armi o armature
		if helpers.InArray(c.Message.Text, []string{
			helpers.Trans("armors"),
			helpers.Trans("weapons"),
		}) {
			c.State.Stage = 1
			return false
		}
	case 1:
		// Verifico se è un delle categorie censite
		if helpers.InArray(c.Message.Text, helpers.GetAllTranslatedSlugCategoriesByLocale()) {
			c.State.Stage = 2
			return false
		}
	case 2:
		// Verifico se è stato richiesto di aggiungere una risorsa
		if strings.Contains(c.Message.Text, helpers.Trans("crafting.add")) {
			c.AddResourceFlag = true
			return false
		}

		// Se è stato richiamato il CRAFT verifico anche che nel peload ci sia almeno una risorsa
		c.Validation.Message = helpers.Trans("crafting.choose_one_resource_to_craft")
		if c.Message.Text == helpers.Trans("crafting.craft") && len(c.Payload.Resources) > 0 {
			c.State.Stage = 3
			return false
		}
	case 3:
		// Se la ricetta viene confermata
		if c.Message.Text == helpers.Trans("confirm") {
			c.State.FinishAt = helpers.GetEndTime(0, 1, 10)
			c.State.Stage = 4
			c.State.ToNotify = helpers.SetTrue()

			c.Validation.Message = helpers.Trans("crafting.wait", c.State.FinishAt.Format("15:04:05"))
			return true
		}
	case 4:
		// Verifico se ha finito il craftin
		c.Validation.Message = helpers.Trans("crafting.wait", c.State.FinishAt.Format("15:04:05"))
		if time.Now().After(c.State.FinishAt) {
			return false
		}
	}

	return true
}

//====================================
// Stage  0 -> 1 - What -> 2 - Category -> 3 - Resources -> 4 - Craft
//====================================
func (c *CraftingController) Stage() {
	var err error

	switch c.State.Stage {
	case 0:
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.what"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("armors")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("weapons")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)
	case 1:
		// Recupero la tipologia di craft scelta dal player
		c.Payload.Item = c.Message.Text

		// Recupero e costruisco tastiera con le categorie in base alla tipologia scelta
		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch c.Payload.Item {
		// ARMORS
		case helpers.Trans("armors"):
			armorCategories, err := providers.GetAllArmorCategory()
			if err != nil {
				services.ErrorHandler("Cant get armor categories", err)
			}

			for _, category := range armorCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(category.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		// WEAPONS
		case helpers.Trans("weapons"):
			weaponCategories, err := providers.GetAllWeaponCategory()
			if err != nil {
				services.ErrorHandler("Cant get weapon categories", err)
			}

			for _, category := range weaponCategories {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(helpers.Trans(category.Slug)))
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Aggiungo anche tasti per uscire
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
		))

		// Invio messaggio con le categorie per la tipologia scelta
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.type"))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		services.SendMessage(msg)

		// Aggiorno stato
		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}
	case 2:
		////////////////////////////////////
		// ONLY FOR DEBUG - Add one resource
		////////////////////////////////////
		_, err := providers.AddResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
			ItemID:   42,
			Quantity: 2,
		})
		if err != nil {
			services.ErrorHandler("Cant add resource to player inventory", err)
		}
		////////////////////////////////////

		// Mi converto per comodità le risorse del player in map
		playerResources := helpers.InventoryToMap(helpers.Player.Inventory)

		// Nel caso in cui l'utente avesse scelto di aggiungere risorse
		// ( Questo stato viene settato in validator )
		if c.AddResourceFlag {
			// Inizializzo, nel caso in cui non fosse la prima aggiunta
			if c.Payload.Resources == nil {
				c.Payload.Resources = make(map[uint]int)
			}

			// Recupero nome della risorsa scelta andandola a pulire da altro testo
			resourceName := strings.Split(
				strings.Split(c.Message.Text, " (")[0],
				helpers.Trans("crafting.add")+" ")[1]

			// Recupero dal WS il dettaglio della risorsa scelta
			resource, err := providers.FindResourceByName(resourceName)
			if err != nil {
				services.ErrorHandler("Cant find resource", err)
			}

			// Mi recupero quante risorse possiede l'utente dalla mappa e controllo che non superi il suo limite
			resourceMaxQuantity := playerResources[resource.ID]
			if helpers.KeyInMap(resource.ID, c.Payload.Resources) {
				if c.Payload.Resources[resource.ID] < resourceMaxQuantity {
					c.Payload.Resources[resource.ID]++
				}
			} else {
				c.Payload.Resources[resource.ID] = 1
			}
		} else {
			c.Payload.Category = helpers.Slugger(c.Message.Text)
		}

		// Ritorno keyboard con la lista delle risorse
		var keyboardRowResources [][]tgbotapi.KeyboardButton
		for r, q := range playerResources {
			// If PayloadResouces < Inventory quantity ok :)
			if c.Payload.Resources[r] < q {
				resource, err := providers.GetResourceByID(r)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.add") + " " + resource.Item.Name + " (" + (strconv.Itoa(q - c.Payload.Resources[r])) + ")",
				))
				keyboardRowResources = append(keyboardRowResources, keyboardRow)
			}
		}

		// Se sono state aggiunte delle risorse nella lista craft aggiungo anche il bottone CRAFT!
		if len(c.Payload.Resources) > 0 {
			keyboardRowResources = append(keyboardRowResources, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans("crafting.craft"),
				),
			))
		}

		// Appendo anche bottone clear o exit
		keyboardRowResources = append(keyboardRowResources,
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)

		//Add recipe message
		var recipe string
		if len(c.Payload.Resources) > 0 {
			for k, v := range c.Payload.Resources {
				resource, err := providers.GetResourceByID(k)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				recipe += resource.Item.Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		// Invio messaggio
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.choose_resources")+"\n"+recipe)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowResources,
		}
		services.SendMessage(msg)

		payloadUpdated, _ := json.Marshal(c.Payload)
		c.State.Payload = string(payloadUpdated)
		c.State, err = providers.UpdatePlayerState(c.State)
		if err != nil {
			services.ErrorHandler("Cant update player", err)
		}
	case 3:
		// Costruisco resoconto ricetta craft
		var recipe string
		if len(c.Payload.Resources) > 0 {
			for k, v := range c.Payload.Resources {
				resource, err := providers.GetResourceByID(k)
				if err != nil {
					services.ErrorHandler("Cant get resource", err)
				}

				recipe += resource.Item.Name + " x " + strconv.Itoa(v) + "\n"
			}
		}

		// Invio messaggio
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.confirm_choose_resources")+"\n\n "+recipe)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.clears")),
			),
		)
		services.SendMessage(msg)
	case 4:
		var craftingResult string

		// Eseguo chiamata al WS in base alla tipologia di craft richiesto
		switch c.Payload.Item {
		case helpers.Trans("armors"):
			// Addatto e costruisco payload
			var craftingRequest nnsdk.ArmorCraft
			helpers.UnmarshalPayload(c.State.Payload, &craftingRequest)

			// Chiamo il WS
			crafted, err := providers.CraftArmor(craftingRequest)
			if err != nil {
				services.ErrorHandler("Cant create armor craft", err)
			}

			// Associo craft al player
			crafted.PlayerID = helpers.Player.ID
			crafted, err = providers.UpdateArmor(crafted)
			if err != nil {
				services.ErrorHandler("Cant associate armor craft", err)
			}

			// For message
			craftingResult = "Name: " + crafted.Name + "\nCategory: " + crafted.ArmorCategory.Name + "\nRarity: " + crafted.Rarity.Name
		case helpers.Trans("weapons"):
			// Addatto e costruisco payload
			var craftingRequest nnsdk.WeaponCraft
			helpers.UnmarshalPayload(c.State.Payload, &craftingRequest)

			// Chiamo il WS
			crafted, err := providers.CraftWeapon(craftingRequest)
			if err != nil {
				services.ErrorHandler("Cant create weapon craft", err)
			}

			// Associo il risultato al player
			crafted.PlayerID = helpers.Player.ID
			crafted, err = providers.UpdateWeapon(crafted)
			if err != nil {
				services.ErrorHandler("Cant associate armor craft", err)
			}

			// For message
			craftingResult = "Name: " + crafted.Name + "\nCategory: " + crafted.WeaponCategory.Name + "\nRarity: " + crafted.Rarity.Name
		}

		// Invio messaggio
		msg := services.NewMessage(c.Message.Chat.ID, helpers.Trans("crafting.craft_completed")+"\n\n"+craftingResult)
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans("route.breaker.back")),
			),
		)
		services.SendMessage(msg)

		// Rimuovo risorse usate al player
		for k, q := range c.Payload.Resources {
			_, err := providers.RemoveResourceToPlayerInventory(helpers.Player, nnsdk.AddResourceRequest{
				ItemID:   k,
				Quantity: q,
			})

			if err != nil {
				services.ErrorHandler("Cant add resource to player inventory", err)
			}
		}

		//====================================
		// COMPLETE!
		//====================================
		helpers.FinishAndCompleteState(c.State, helpers.Player)
		//====================================
	}
}
