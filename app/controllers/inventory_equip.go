package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/app/helpers"
	"bitbucket.org/no-name-game/nn-telegram/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// InventoryEquipController
// ====================================
type InventoryEquipController struct {
	BaseController
	Payload struct {
		Type    string
		EquipID uint
	}
}

// ====================================
// Handle
// ====================================
func (c *InventoryEquipController) Handle(player nnsdk.Player, update tgbotapi.Update) {
	// Inizializzo variabili del controler
	var err error
	c.Controller = "route.inventory.equip"
	c.Player = player
	c.Update = update

	// Verifico lo stato della player
	c.State, _, err = helpers.CheckState(player, c.Controller, c.Payload, c.Father)
	// Se non sono riuscito a recuperare/creare lo stato esplodo male, qualcosa è andato storto.
	if err != nil {
		panic(err)
	}

	// Set and load payload
	helpers.UnmarshalPayload(c.State.Payload, &c.Payload)

	// Validate
	var hasError bool
	hasError, err = c.Validator()
	if err != nil {
		panic(err)
	}

	// Se ritornano degli errori
	if hasError == true {
		// Invio il messaggio in caso di errore e chiudo
		validatorMsg := services.NewMessage(c.Update.Message.Chat.ID, c.Validation.Message)
		validatorMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, "route.breaker.back"),
				),
			),
		)

		_, err = services.SendMessage(validatorMsg)
		if err != nil {
			panic(err)
		}

		return
	}

	// Ok! Run!
	err = c.Stage()
	if err != nil {
		panic(err)
	}

	// Aggiorno stato finale
	payloadUpdated, _ := json.Marshal(c.Payload)
	c.State.Payload = string(payloadUpdated)
	_, err = providers.UpdatePlayerState(c.State)
	if err != nil {
		panic(err)
	}

	// Verifico se lo stato è completato chiudo
	if *c.State.Completed == true {
		_, err = providers.DeletePlayerState(c.State) // Delete
		if err != nil {
			panic(err)
		}

		err = helpers.DelRedisState(player)
		if err != nil {
			panic(err)
		}

		// Call menu controller
		new(MenuController).Handle(c.Player, c.Update)
	}

	return
}

// ====================================
// Validator
// ====================================
func (c *InventoryEquipController) Validator() (hasErrors bool, err error) {
	c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.general")

	switch c.State.Stage {
	// È il primo stato non c'è nessun controllo
	case 0:
		return false, err

	// Verifico che la tipologia di equip che vuole il player esista
	case 1:
		if helpers.InArray(c.Update.Message.Text, []string{
			helpers.Trans(c.Player.Language.Slug, "armors"),
			helpers.Trans(c.Player.Language.Slug, "weapons"),
		}) {
			return false, err
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true, err

	// Verifico che il player voglia continuare con l'equip
	case 2:
		if strings.Contains(c.Update.Message.Text, helpers.Trans(c.Player.Language.Slug, "equip")) {
			return false, err
		}
		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true, err

	// Verifico la conferma dell'equip
	case 3:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "confirm") {
			return false, err
		}

		c.Validation.Message = helpers.Trans(c.Player.Language.Slug, "validator.not_valid")

		return true, err
	}

	return true, err
}

// ====================================
// Stage
// ====================================
func (c *InventoryEquipController) Stage() (err error) {
	switch c.State.Stage {
	// In questo stage faccio un micro recap al player del suo equipaggiamento
	// attuale e mostro a tastierino quale categoria vorrebbe equipaggiare
	case 0:
		var currentPlayerEquipment string
		currentPlayerEquipment = helpers.Trans(c.Player.Language.Slug, "inventory.equip.equipped")

		// ******************
		// Recupero armatura equipaggiata
		// ******************
		var currentArmorsEquipment string
		currentArmorsEquipment = fmt.Sprintf("%s:\n", helpers.Trans(c.Player.Language.Slug, "armors"))

		var armors nnsdk.Armors
		armors, err = providers.GetPlayerArmors(c.Player, "true")
		if err != nil {
			return err
		}

		for _, armor := range armors {
			currentArmorsEquipment += fmt.Sprintf("- %s \n", armor.Name)
		}

		// ******************
		// Recupero armi equipaggiate
		// ******************
		var currentWeaponsEquipment string
		currentWeaponsEquipment = fmt.Sprintf("%s:\n", helpers.Trans(c.Player.Language.Slug, "weapons"))

		var weapons nnsdk.Weapons
		weapons, err := providers.GetPlayerWeapons(c.Player, "true")
		if err != nil {
			return err
		}

		for _, weapon := range weapons {
			currentWeaponsEquipment += fmt.Sprintf("- %s \n", weapon.Name)
		}

		// Invio messagio con recap e con selettore categoria
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s \n %s",
				helpers.Trans(c.Player.Language.Slug, "inventory.type"),
				fmt.Sprintf("%s \n %s \n %s", currentPlayerEquipment, currentArmorsEquipment, currentWeaponsEquipment),
			),
		)

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "armors")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "weapons")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Avanzo di stage
		c.State.Stage = 1

	// In questo stage chiedo di indicarmi quale armatura o arma intende equipaggiare
	case 1:
		// Costruisco keyboard risposta
		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		c.Payload.Type = c.Update.Message.Text

		switch c.Payload.Type {
		case helpers.Trans(c.Player.Language.Slug, "armors"):
			// Recupero nuovamente armature player, richiamando la rotta dedicata
			// in questa maniera posso filtrare per quelle che non sono equipaggiate
			var armors nnsdk.Armors
			armors, err = providers.GetPlayerArmors(c.Player, "false")
			if err != nil {
				return err
			}

			// Ciclo armature del player
			for _, armor := range armors {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						fmt.Sprintf(
							"%s %s",
							helpers.Trans(c.Player.Language.Slug, "equip"),
							armor.Name,
						),
					),
				)
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}

		case helpers.Trans(c.Player.Language.Slug, "weapons"):
			// Recupero nuovamente armi player, richiamando la rotta dedicata
			// in questa maniera posso filtrare per quelle che non sono equipaggiate
			var weapons nnsdk.Weapons
			weapons, err := providers.GetPlayerWeapons(c.Player, "false")
			if err != nil {
				return err
			}

			// Ciclo armi player
			for _, weapon := range weapons {
				keyboardRow := tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						fmt.Sprintf(
							"%s %s",
							helpers.Trans(c.Player.Language.Slug, "equip"),
							weapon.Name,
						),
					),
				)
				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Aggiungo tasti back and clears
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
		))

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "inventory.what"))
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.State.Stage = 2

	// In questo stato ricerco effettivamente l'arma o l'armatura che il player vuole
	// equipaggiare e me lo metto nel payload in attesa di conferma
	case 2:
		var equipmentName string
		var equipmentID uint

		// Ripulisco messaggio per recupermi solo il nome
		equipmentName = strings.Split(c.Update.Message.Text, helpers.Trans(c.Player.Language.Slug, "equip")+" ")[1]

		switch c.Payload.Type {
		case helpers.Trans(c.Player.Language.Slug, "armors"):
			var armor nnsdk.Armor
			armor, err := providers.FindArmorByName(equipmentName)
			if err != nil {
				return err
			}

			equipmentID = armor.ID
		case helpers.Trans(c.Player.Language.Slug, "weapons"):
			var weapon nnsdk.Weapon
			weapon, err := providers.FindWeaponByName(equipmentName)
			if err != nil {
				return err
			}

			equipmentID = weapon.ID
		}

		// Invio messaggio per conferma equipaggiamento
		msg := services.NewMessage(c.Update.Message.Chat.ID,
			fmt.Sprintf(
				"%s \n\n %s",
				helpers.Trans(c.Player.Language.Slug, "inventory.equip.confirm"),
				equipmentName,
			),
		)

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.clears")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Aggiorno stato
		c.Payload.EquipID = equipmentID
		c.State.Stage = 3

	// In questo stage se l'utente ha confermato continuo con l'equipaggiamento
	// TODO: bisogna verifica che ci sia solo 1 arma o armatura equipaggiata
	case 3:
		switch c.Payload.Type {
		case helpers.Trans(c.Player.Language.Slug, "armors"):
			equipment, err := providers.GetArmorByID(c.Payload.EquipID)
			if err != nil {
				return err
			}

			// Aggiorno equipped
			equipment.Equipped = helpers.SetTrue()
			_, err = providers.UpdateArmor(equipment)
			if err != nil {
				return err
			}

		case helpers.Trans(c.Player.Language.Slug, "weapons"):
			equipment, err := providers.GetWeaponByID(c.Payload.EquipID)
			if err != nil {
				return err
			}

			// Aggiorno equipped
			equipment.Equipped = helpers.SetTrue()
			_, err = providers.UpdateWeapon(equipment)
			if err != nil {
				return err
			}
		}

		// Invio messaggio
		msg := services.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "inventory.equip.completed"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		_, err = services.SendMessage(msg)
		if err != nil {
			return err
		}

		// Completo lo stato
		c.State.Completed = helpers.SetTrue()
	}

	return
}
