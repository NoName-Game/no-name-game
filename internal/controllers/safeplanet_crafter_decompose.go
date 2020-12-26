package controllers

import (
	"fmt"
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetCrafterDecomposeController
// ====================================
type SafePlanetCrafterDecomposeController struct {
	Payload struct {
		ItemType string
		ItemID   uint32
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetCrafterDecomposeController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.crafter.decompose",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCrafterController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
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
func (c *SafePlanetCrafterDecomposeController) Validator() bool {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico tipologia item che il player vuole craftare
	// ##################################################################################################
	case 0:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "armor") {
			c.Payload.ItemType = "armor"
			c.CurrentState.Stage = 1
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "weapon") {
			c.Payload.ItemType = "weapon"
			c.CurrentState.Stage = 1
		}

	// ##################################################################################################
	// Verifico se l'arma scelta esiste ed è sua
	// ##################################################################################################
	case 2:
		playerChoice := strings.Split(c.Update.Message.Text, " (")[0]

		// TODO: differenziare in base alla tipologia

		switch c.Payload.ItemType {
		case "armor":
			var rGetPlayerArmors *pb.GetPlayerArmorsResponse
			rGetPlayerArmors, _ = config.App.Server.Connection.GetPlayerArmors(helpers.NewContext(1), &pb.GetPlayerArmorsRequest{
				PlayerID: c.Player.ID,
			})

			for _, armor := range rGetPlayerArmors.GetArmors() {
				if armor.GetName() == playerChoice {
					c.Payload.ItemID = armor.GetID()
					return false
				}
			}
		case "weapon":
			var rGetPlayerWeapons *pb.GetPlayerWeaponsResponse
			rGetPlayerWeapons, _ = config.App.Server.Connection.GetPlayerWeapons(helpers.NewContext(1), &pb.GetPlayerWeaponsRequest{
				PlayerID: c.Player.ID,
			})

			for _, weapon := range rGetPlayerWeapons.GetWeapons() {
				if weapon.GetName() == playerChoice {
					c.Payload.ItemID = weapon.GetID()
					return false
				}
			}
		}

		return true
	// ##################################################################################################
	// Verifico Conferma
	// ##################################################################################################
	case 3:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetCrafterDecomposeController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	case 0:
		startMsg := fmt.Sprintf("%s\n\n%s",
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.decompose.info"),
			helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.decompose.what"),
		)

		msg := helpers.NewMessage(c.Player.ChatID, startMsg)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "armor")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "weapon")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		// Invio messaggio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

	// ##################################################################################################
	// Chiedo al player di indicare cosa vuole riciclare
	// ##################################################################################################
	case 1:
		var rowsKeyboard [][]tgbotapi.KeyboardButton

		switch c.Payload.ItemType {
		case "armor":
			// Recupero armature player
			var rGetPlayerArmors *pb.GetPlayerArmorsResponse
			rGetPlayerArmors, _ = config.App.Server.Connection.GetPlayerArmors(helpers.NewContext(1), &pb.GetPlayerArmorsRequest{
				PlayerID: c.Player.ID,
			})

			for _, armor := range rGetPlayerArmors.GetArmors() {
				rowsKeyboard = append(rowsKeyboard, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						fmt.Sprintf("%s (%s)",
							armor.GetName(),
							armor.GetRarity().GetSlug(),
						),
					),
				))
			}
		case "weapon":
			// Recupero armi player
			var rGetPlayerWeapons *pb.GetPlayerWeaponsResponse
			rGetPlayerWeapons, _ = config.App.Server.Connection.GetPlayerWeapons(helpers.NewContext(1), &pb.GetPlayerWeaponsRequest{
				PlayerID: c.Player.ID,
			})

			for _, weapon := range rGetPlayerWeapons.GetWeapons() {
				rowsKeyboard = append(rowsKeyboard, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						fmt.Sprintf("%s (%s)",
							weapon.GetName(),
							weapon.GetRarity().GetSlug(),
						),
					),
				))
			}
		}

		rowsKeyboard = append(rowsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		))

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.decompose.which"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       rowsKeyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 2
	// ##################################################################################################
	// Chiedo Conferma al player
	// ##################################################################################################
	case 2:
		var itemDetails string
		switch c.Payload.ItemType {
		case "armor":
			// Recupero dettagli armatura
			var rGetArmorByID *pb.GetArmorByIDResponse
			if rGetArmorByID, err = config.App.Server.Connection.GetArmorByID(helpers.NewContext(1), &pb.GetArmorByIDRequest{
				ArmorID: c.Payload.ItemID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			itemDetails = fmt.Sprintf("%s (%s)", rGetArmorByID.GetArmor().GetName(), rGetArmorByID.GetArmor().GetRarity().GetSlug())
		case "weapon":
			// Recupero dettagli arma
			var rGetWeaponByID *pb.GetWeaponByIDResponse
			if rGetWeaponByID, err = config.App.Server.Connection.GetWeaponByID(helpers.NewContext(1), &pb.GetWeaponByIDRequest{
				ID: c.Payload.ItemID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			itemDetails = fmt.Sprintf("%s (%s)", rGetWeaponByID.GetWeapon().GetName(), rGetWeaponByID.GetWeapon().GetRarity().GetSlug())
		}

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.decompose.confirm",
			itemDetails,
		))

		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 3
	// ##################################################################################################
	// Riciclo
	// ##################################################################################################
	case 3:
		var pbEquipType pb.CrafterDecomposeRequest_EquipTypeEnum
		switch c.Payload.ItemType {
		case "weapon":
			pbEquipType = pb.CrafterDecomposeRequest_WEAPON
		case "armor":
			pbEquipType = pb.CrafterDecomposeRequest_ARMOR
		}

		// Avvio scomponi equipaggiamento
		var rCrafterDecompose *pb.CrafterDecomposeResponse
		rCrafterDecompose, err = config.App.Server.Connection.CrafterDecompose(helpers.NewContext(1), &pb.CrafterDecomposeRequest{
			PlayerID:  c.Player.ID,
			EquipType: pbEquipType,
			EquipID:   c.Payload.ItemID,
		})

		//TODO: da completare controllo monete
		if err != nil && strings.Contains(err.Error(), "player dont have enough money") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di monete
			errorMsg := helpers.NewMessage(c.Update.Message.Chat.ID,
				helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.no_money"),
			)
			if _, err = helpers.SendMessage(errorMsg); err != nil {
				c.Logger.Panic(err)
			}

			c.CurrentState.Completed = true
			return
		} else if err != nil {
			c.Logger.Panic(err)
		}

		// Recap risorse recuperate
		var recap string
		for _, resource := range rCrafterDecompose.GetResources() {
			recap += fmt.Sprintf("- %s (%s)\n", resource.GetName(), resource.GetRarity().GetSlug())
		}

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.decompose.completed", recap))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Completed = true
	}
}
