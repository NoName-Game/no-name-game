package controllers

import (
	"fmt"
	"math"
	"strings"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// PlayerEquipmentController
// ====================================
type PlayerEquipmentController struct {
	Controller
	Payload struct {
		ItemType     string // Armor/Weapon
		ItemCategory string // Head/Leg/Chest/Arms
		EquipID      uint32
	}
}

// ====================================
// Handle
// ====================================
func (c *PlayerEquipmentController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.player.inventory.equip",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &PlayerController{},
				FromStage: 1,
			},
			BreakerPerStage: map[int32][]string{
				1: {"route.breaker.menu"},
				2: {"route.breaker.menu", "route.breaker.back"},
				3: {"route.breaker.menu", "route.breaker.back"},
				4: {"route.breaker.menu", "route.breaker.back"},
			},
		},
	}) {
		return
	}

	// Validate
	if c.Validator() {
		c.Validate()
		return
	}

	// Ok! Run!
	c.Stage()

	// Completo progressione
	c.Completing(&c.Payload)
}

// ====================================
// Validator
// ====================================
func (c *PlayerEquipmentController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico quale tipologia di equipaggiamento si vuole gestire
	// ##################################################################################################
	case 1:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "armors") {
			c.Payload.ItemType = "armors"
			return false
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "weapons") {
			// Se viene richiesto di craftare un'arma passo direttamente alla lista delle armi
			// in quanto le armi non hanno una categoria
			c.Payload.ItemType = "weapons"
			c.CurrentState.Stage = 2
			return false
		}

		return true
	// ##################################################################################################
	// Verifico che la tipologia di equip che vuole il player esista
	// ##################################################################################################
	case 2:
		// Recupero categoria scelta in base alla formattazione passata
		categorySplit := strings.Split(c.Update.Message.Text, " (")
		if c.Payload.ItemCategory = helpers.CheckAndReturnCategorySlug(c.Player.Language.Slug, categorySplit[0]); c.Payload.ItemCategory == "" {
			return true
		}
	// ##################################################################################################
	// Verifico conferma equipaggiamento player
	// ##################################################################################################
	case 3:
		if strings.Contains(c.Update.Message.Text, "🗡") || strings.Contains(c.Update.Message.Text, "🛡") {
			return false
		}

		return true
	// ##################################################################################################
	// Verifico conferma equipaggiamento player
	// ##################################################################################################
	case 4:
		// Verifico la conferma dell'equip
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *PlayerEquipmentController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// In questo stage faccio un micro recap al player del suo equipaggiamento
	// attuale e mostro a tastierino quale categoria vorrebbe equipaggiare
	case 0:
		// ******************
		// Recupero armatura equipaggiata
		// ******************

		var currentArmorsEquipment string
		currentArmorsEquipment = fmt.Sprintf("<b>%s</b>:", helpers.Trans(c.Player.Language.Slug, "armor"))

		var rGetPlayerArmors *pb.GetPlayerArmorsEquippedResponse
		if rGetPlayerArmors, err = config.App.Server.Connection.GetPlayerArmorsEquipped(helpers.NewContext(1), &pb.GetPlayerArmorsEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// armatura base player
		if len(rGetPlayerArmors.GetArmors()) > 0 {
			// var head, gauntlets, chest, leg string
			armors := helpers.SortPlayerArmor(rGetPlayerArmors.GetArmors())

			for _, armor := range armors {
				if armor != nil {
					currentArmorsEquipment += fmt.Sprintf("\n%s [<b>%s</b>] (%s)\nDEF: <b>%v</b> | EVS: <b>%v</b> | HLV: <b>%v</b>\n",
						helpers.Trans(c.Player.Language.Slug, armor.GetArmorCategory().GetSlug()+"_emoji"),
						armor.Name, strings.ToUpper(armor.GetRarity().GetSlug()),
						math.Round(armor.Defense),
						math.Round(armor.Evasion),
						math.Round(armor.Halving),
					)
				}
			}
		} else {
			currentArmorsEquipment += helpers.Trans(c.Player.Language.Slug, "inventory.armors.zero_equipment")
		}

		// ******************
		// Recupero arma equipaggiata
		// ******************
		var currentWeaponsEquipment string
		currentWeaponsEquipment = fmt.Sprintf("<b>%s</b>:", helpers.Trans(c.Player.Language.Slug, "weapon"))

		var rGetPlayerWeaponEquippedResponse *pb.GetPlayerWeaponEquippedResponse
		if rGetPlayerWeaponEquippedResponse, err = config.App.Server.Connection.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
			PlayerID: c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		if rGetPlayerWeaponEquippedResponse.GetWeapon() != nil {
			currentWeaponsEquipment += fmt.Sprintf(
				"\n<b>%s</b> (<b>%s</b>)\nDamage: <b>%v</b> | Precision: <b>%v</b>",
				rGetPlayerWeaponEquippedResponse.GetWeapon().GetName(),
				strings.ToUpper(rGetPlayerWeaponEquippedResponse.GetWeapon().GetRarity().GetSlug()),
				math.Round(rGetPlayerWeaponEquippedResponse.GetWeapon().GetRawDamage()),
				math.Round(rGetPlayerWeaponEquippedResponse.GetWeapon().GetPrecision()),
			)
		} else {
			currentWeaponsEquipment += helpers.Trans(c.Player.Language.Slug, "inventory.weapons.zero_equipment")
		}

		// Invio messagio con recap e con selettore categoria
		msg := helpers.NewMessage(c.ChatID,
			fmt.Sprintf(
				"%s\n\n%s\n%s",
				helpers.Trans(c.Player.Language.Slug, "inventory.equip.equipped"),
				currentArmorsEquipment,
				currentWeaponsEquipment,
			),
		)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "armors")),
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "weapons")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Avanzo di stage
		c.CurrentState.Stage = 1

	// In questo stage chiedo di indicarmi quale armatura o arma intende equipaggiare
	case 1:
		var message string
		var keyboardRowCategories [][]tgbotapi.KeyboardButton

		switch c.Payload.ItemType {
		case "armors":
			message = helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.armor.type")

			var rGetAllArmorCategory *pb.GetAllArmorCategoryResponse
			if rGetAllArmorCategory, err = config.App.Server.Connection.GetAllArmorCategory(helpers.NewContext(1), &pb.GetAllArmorCategoryRequest{}); err != nil {
				return
			}

			for _, category := range rGetAllArmorCategory.GetArmorCategories() {
				// Recupero quante armature per quella tipologia del player per contare quante ne ha
				var rGetPlayerArmorsByCategoryID *pb.GetPlayerArmorsByCategoryIDResponse
				if rGetPlayerArmorsByCategoryID, err = config.App.Server.Connection.GetPlayerArmorsByCategoryID(helpers.NewContext(1), &pb.GetPlayerArmorsByCategoryIDRequest{
					PlayerID:   c.Player.GetID(),
					CategoryID: category.GetID(),
				}); err != nil {
					c.Logger.Panic(err)
				}

				keyboardRow := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(
					fmt.Sprintf("%s (%v)",
						helpers.Trans(c.Player.Language.Slug, category.Slug),
						len(rGetPlayerArmorsByCategoryID.GetArmors()),
					),
				))

				keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
			}
		}

		// Clear and exit
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		msg := helpers.NewMessage(c.Player.ChatID, message)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		var mainMessage string

		// Costruisco keyboard risposta
		var keyboardRowCategories [][]tgbotapi.KeyboardButton
		switch c.Payload.ItemType {
		// **************************
		// GESTIONE ARMATURE
		// **************************
		case "armors":
			mainMessage = helpers.Trans(c.Player.Language.Slug, "inventory.armors.no_one")

			// Recupero categoria scelta dal player
			var rGetArmorCategoryBySlugRequest *pb.GetArmorCategoryBySlugResponse
			if rGetArmorCategoryBySlugRequest, err = config.App.Server.Connection.GetArmorCategoryBySlug(helpers.NewContext(1), &pb.GetArmorCategoryBySlugRequest{
				Slug: c.Payload.ItemCategory,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Recupero armature non equipaggiate filtrate per la categoria scelta
			var rGetPlayerArmorsByCategoryID *pb.GetPlayerArmorsByCategoryIDResponse
			if rGetPlayerArmorsByCategoryID, err = config.App.Server.Connection.GetPlayerArmorsByCategoryID(helpers.NewContext(1), &pb.GetPlayerArmorsByCategoryIDRequest{
				PlayerID:   c.Player.GetID(),
				CategoryID: rGetArmorCategoryBySlugRequest.GetArmorCategory().GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			if len(rGetPlayerArmorsByCategoryID.GetArmors()) > 0 {
				mainMessage = helpers.Trans(c.Player.Language.Slug, "inventory.armors.what")

				// Ciclo armature del player
				for _, armor := range rGetPlayerArmorsByCategoryID.GetArmors() {
					mainMessage += fmt.Sprintf(
						"\n<b>%s</b> (%s) 🛡 [%v, %v%%, %v%%] 🎖%v",
						armor.Name,
						strings.ToUpper(armor.Rarity.Slug),
						math.Round(armor.Defense),
						math.Round(armor.Evasion),
						math.Round(armor.Halving),
						armor.Rarity.LevelToEuip,
					)
					keyboardRow := tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							fmt.Sprintf(
								"%s (%s) 🛡 [%v, %v%%, %v%%] 🎖%v",
								// helpers.Trans(c.Player.Language.Slug, "equip"),
								armor.Name,
								strings.ToUpper(armor.Rarity.Slug),
								math.Round(armor.Defense),
								math.Round(armor.Evasion),
								math.Round(armor.Halving),
								armor.Rarity.LevelToEuip,
							),
						),
					)
					keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
				}
			}

		// **************************
		// GESTIONE ARMI
		// **************************
		case "weapons":
			mainMessage = helpers.Trans(c.Player.Language.Slug, "inventory.armors.no_one")

			// Recupero nuovamente armi player, richiamando la rotta dedicata
			// in questa maniera posso filtrare per quelle che non sono equipaggiate
			var rGetPlayerWeapons *pb.GetPlayerWeaponsResponse
			if rGetPlayerWeapons, err = config.App.Server.Connection.GetPlayerWeapons(helpers.NewContext(1), &pb.GetPlayerWeaponsRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			if len(rGetPlayerWeapons.GetWeapons()) > 0 {
				mainMessage = helpers.Trans(c.Player.Language.Slug, "inventory.weapons.what")

				// Ciclo armi player
				for _, weapon := range rGetPlayerWeapons.GetWeapons() {
					mainMessage += fmt.Sprintf(
						"\n<b>(%s)</b> (%s) 🗡 [%v, %v%%, %v] 🎖%v",
						weapon.Name,
						strings.ToUpper(weapon.Rarity.Slug),
						math.Round(weapon.RawDamage),
						math.Round(weapon.Precision),
						weapon.Durability,
						weapon.Rarity.LevelToEuip,
					)
					keyboardRow := tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							fmt.Sprintf(
								"%s (%s) 🗡 [%v, %v%%, %v] 🎖%v",
								weapon.Name,
								strings.ToUpper(weapon.Rarity.Slug),
								math.Round(weapon.RawDamage),
								math.Round(weapon.Precision),
								weapon.Durability,
								weapon.Rarity.LevelToEuip,
							),
						),
					)
					keyboardRowCategories = append(keyboardRowCategories, keyboardRow)
				}
			}
		}

		// Aggiungo tasti back and clears
		keyboardRowCategories = append(keyboardRowCategories, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID, mainMessage)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRowCategories,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3

	// In questo stato ricerco effettivamente l'arma o l'armatura che il player vuole
	// equipaggiare e me lo metto nel payload in attesa di conferma
	case 3:
		var equipmentName string
		var equipmentID uint32
		var equipmentError bool
		var confirmMessage string

		// Ripulisco messaggio per recupermi solo il nome
		equipmentName = strings.Split(c.Update.Message.Text, " (")[0]
		switch c.Payload.ItemType {
		case "armors":
			var rGetArmorByName *pb.GetArmorByPlayerAndNameResponse
			if rGetArmorByName, err = config.App.Server.Connection.GetArmorByPlayerAndName(helpers.NewContext(1), &pb.GetArmorByPlayerAndNameRequest{
				Name:     equipmentName,
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Verifico se appartiene correttamente al player
			if rGetArmorByName.GetArmor().GetPlayerID() != c.Player.ID {
				equipmentError = true
			}
			equipmentID = rGetArmorByName.GetArmor().GetID()

			// Verifico che il player possieda il livello necessario per equipaggarla
			if rGetArmorByName.GetArmor().GetRarity().GetLevelToEuip() > int32(c.Player.GetLevelID()) {
				errorMsg := helpers.NewMessage(c.ChatID,
					helpers.Trans(c.Player.Language.Slug, "inventory.equip.level_lower", rGetArmorByName.GetArmor().GetRarity().GetLevelToEuip()),
				)
				errorMsg.ParseMode = tgbotapi.ModeHTML
				if _, err = helpers.SendMessage(errorMsg); err != nil {
					c.Logger.Panic(err)
				}
				return
			}

			// Recupero armatura attualmente equipaggiata per la categoria scelta
			var rGetPlayerArmorEquippedByCategoryID *pb.GetPlayerArmorEquippedByCategoryIDResponse
			rGetPlayerArmorEquippedByCategoryID, _ = config.App.Server.Connection.GetPlayerArmorEquippedByCategoryID(helpers.NewContext(1), &pb.GetPlayerArmorEquippedByCategoryIDRequest{
				PlayerID:   c.Player.GetID(),
				CategoryID: rGetArmorByName.GetArmor().GetArmorCategory().GetID(),
			})

			// Preparo messaggio di conferma
			confirmMessage = helpers.Trans(c.Player.Language.Slug, "inventory.equip.confirm", rGetArmorByName.GetArmor().GetName())

			// Se ha un'armaatura equipaggiata chiedo switch
			if rGetPlayerArmorEquippedByCategoryID.GetArmor().GetID() > 0 {
				confirmMessage = helpers.Trans(c.Player.Language.Slug, "inventory.equip.confirm_armor",
					rGetArmorByName.GetArmor().GetName(),
					rGetPlayerArmorEquippedByCategoryID.GetArmor().GetName(),
				)
			}
		case "weapons":
			var rGetWeaponByName *pb.GetWeaponByPlayerAndNameResponse
			if rGetWeaponByName, err = config.App.Server.Connection.GetWeaponByPlayerAndName(helpers.NewContext(1), &pb.GetWeaponByPlayerAndNameRequest{
				Name:     equipmentName,
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Verifico se appartiene correttamente al player
			if rGetWeaponByName.GetWeapon().GetPlayerID() != c.Player.ID {
				equipmentError = true
			}

			equipmentID = rGetWeaponByName.GetWeapon().GetID()

			// Verifico che il player possieda il livello necessario per equipaggarla
			if rGetWeaponByName.GetWeapon().GetRarity().GetLevelToEuip() > int32(c.Player.GetLevelID()) {
				errorMsg := helpers.NewMessage(c.ChatID,
					helpers.Trans(c.Player.Language.Slug, "inventory.equip.level_lower", rGetWeaponByName.GetWeapon().GetRarity().GetLevelToEuip()),
				)
				errorMsg.ParseMode = tgbotapi.ModeHTML
				if _, err = helpers.SendMessage(errorMsg); err != nil {
					c.Logger.Panic(err)
				}
				return
			}

			// Recupero arma attualmente equipaggiata
			var rGetPlayerWeaponEquipped *pb.GetPlayerWeaponEquippedResponse
			rGetPlayerWeaponEquipped, _ = config.App.Server.Connection.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
				PlayerID: c.Player.GetID(),
			})

			// Preparo messaggio di conferma
			confirmMessage = helpers.Trans(c.Player.Language.Slug, "inventory.equip.confirm", rGetWeaponByName.GetWeapon().GetName())

			// Se ha un'arma equipaggiata chiedo switch
			if rGetPlayerWeaponEquipped.GetWeapon().GetID() > 0 {
				confirmMessage = helpers.Trans(c.Player.Language.Slug, "inventory.equip.confirm_weapon",
					rGetWeaponByName.GetWeapon().GetName(),
					rGetPlayerWeaponEquipped.GetWeapon().GetName(),
				)
			}

		}

		if equipmentError {
			// Invio messaggio error
			msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "inventory.equip.error"))
			msg.ParseMode = tgbotapi.ModeHTML
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
				),
			)

			if _, err = helpers.SendMessage(msg); err != nil {
				c.Logger.Panic(err)
			}

			return
		}

		// Invio messaggio per conferma equipaggiamento
		msg := helpers.NewMessage(c.ChatID, confirmMessage)
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.Payload.EquipID = equipmentID
		c.CurrentState.Stage = 4

	// In questo stage se l'utente ha confermato continuo con l'equipaggiamento
	case 4:
		switch c.Payload.ItemType {
		case "armors":
			// Recupero categoria scelta dal player
			var rGetArmorCategoryBySlugRequest *pb.GetArmorCategoryBySlugResponse
			if rGetArmorCategoryBySlugRequest, err = config.App.Server.Connection.GetArmorCategoryBySlug(helpers.NewContext(1), &pb.GetArmorCategoryBySlugRequest{
				Slug: c.Payload.ItemCategory,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Recupero armatura attualmente equipaggiata per la categoria scelta
			var rGetPlayerArmorEquippedByCategoryID *pb.GetPlayerArmorEquippedByCategoryIDResponse
			rGetPlayerArmorEquippedByCategoryID, _ = config.App.Server.Connection.GetPlayerArmorEquippedByCategoryID(helpers.NewContext(1), &pb.GetPlayerArmorEquippedByCategoryIDRequest{
				PlayerID:   c.Player.GetID(),
				CategoryID: rGetArmorCategoryBySlugRequest.GetArmorCategory().GetID(),
			})

			// Rimuovo equipaggiamento attuale
			if rGetPlayerArmorEquippedByCategoryID.GetArmor().GetID() > 0 {
				if _, err = config.App.Server.Connection.EquipArmor(helpers.NewContext(1), &pb.EquipArmorRequest{
					PlayerID: c.Player.ID,
					ArmorID:  rGetPlayerArmorEquippedByCategoryID.GetArmor().GetID(),
					Equip:    false,
				}); err != nil {
					c.Logger.Panic(err)
				}
			}

			// Aggiorno con quello nuovo
			_, err = config.App.Server.Connection.EquipArmor(helpers.NewContext(1), &pb.EquipArmorRequest{
				PlayerID: c.Player.ID,
				ArmorID:  c.Payload.EquipID,
				Equip:    true,
			})

			if err != nil && strings.Contains(err.Error(), "level to equip is higher then player level") {
				// Recupero dettagli arma
				var rGetArmorByID *pb.GetArmorByIDResponse
				if rGetArmorByID, err = config.App.Server.Connection.GetArmorByID(helpers.NewContext(1), &pb.GetArmorByIDRequest{
					ArmorID: c.Payload.EquipID,
				}); err != nil {
					c.Logger.Panic(err)
				}

				// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
				errorMsg := helpers.NewMessage(c.ChatID,
					helpers.Trans(c.Player.Language.Slug, "inventory.equip.level_lower", rGetArmorByID.GetArmor().GetRarity().GetLevelToEuip()),
				)
				errorMsg.ParseMode = tgbotapi.ModeHTML
				if _, err = helpers.SendMessage(errorMsg); err != nil {
					c.Logger.Panic(err)
				}
				return
			} else if err != nil {
				c.Logger.Panic(err)
			}

		case "weapons":
			// Recupero arma attualmente equipaggiata
			var rGetPlayerWeaponEquipped *pb.GetPlayerWeaponEquippedResponse
			rGetPlayerWeaponEquipped, _ = config.App.Server.Connection.GetPlayerWeaponEquipped(helpers.NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
				PlayerID: c.Player.GetID(),
			})

			// Rimovo arma attualmente equipaggiata
			if rGetPlayerWeaponEquipped.GetWeapon().GetID() > 0 {
				if _, err = config.App.Server.Connection.EquipWeapon(helpers.NewContext(1), &pb.EquipWeaponRequest{
					PlayerID: c.Player.ID,
					WeaponID: rGetPlayerWeaponEquipped.GetWeapon().GetID(),
					Equip:    false,
				}); err != nil {
					c.Logger.Panic(err)
				}
			}

			// Aggiorno equipped
			_, err = config.App.Server.Connection.EquipWeapon(helpers.NewContext(1), &pb.EquipWeaponRequest{
				PlayerID: c.Player.ID,
				WeaponID: c.Payload.EquipID,
				Equip:    true,
			})

			if err != nil && strings.Contains(err.Error(), "level to equip is higher then player level") {
				// Recupero dettagli arma
				var rGetWeaponByID *pb.GetWeaponByIDResponse
				if rGetWeaponByID, err = config.App.Server.Connection.GetWeaponByID(helpers.NewContext(1), &pb.GetWeaponByIDRequest{
					ID: c.Payload.EquipID,
				}); err != nil {
					c.Logger.Panic(err)
				}

				// Potrebbero esserci stati degli errori come per esempio la mancanza di materie prime
				errorMsg := helpers.NewMessage(c.ChatID,
					helpers.Trans(c.Player.Language.Slug, "inventory.equip.level_lower", rGetWeaponByID.GetWeapon().GetRarity().GetLevelToEuip()),
				)
				errorMsg.ParseMode = tgbotapi.ModeHTML
				if _, err = helpers.SendMessage(errorMsg); err != nil {
					c.Logger.Panic(err)
				}
				return
			} else if err != nil {
				c.Logger.Panic(err)
			}
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.ChatID,
			helpers.Trans(c.Player.Language.Slug, "inventory.equip.completed"),
		)
		msg.ParseMode = tgbotapi.ModeHTML

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true

		// ###################
		// TUTORIAL - Solo il player si trova dentro il tutorial forzo di tornarare al menu
		// ###################
		if c.InTutorial() {
			c.Configurations.ControllerBack.To = &MenuController{}
		}
	}

	return
}
