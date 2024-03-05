package controllers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"
)

// ====================================
// SafePlanetCrafterRepairController
// ====================================
type SafePlanetCrafterRepairController struct {
	Payload struct {
		WeaponID uint32
	}
	Controller
}

func (c *SafePlanetCrafterRepairController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.crafter.repair",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetCrafterController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
				2: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetCrafterRepairController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
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
func (c *SafePlanetCrafterRepairController) Validator() bool {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se l'arma scelta esiste ed è sua
	// ##################################################################################################
	case 1:
		weaponMsg := strings.Split(c.Update.Message.Text, " -")[0]

		var rGetPlayerWeapons *pb.GetPlayerWeaponsResponse
		rGetPlayerWeapons, _ = config.App.Server.Connection.GetPlayerWeapons(helpers.NewContext(1), &pb.GetPlayerWeaponsRequest{
			PlayerID: c.Player.ID,
		})

		for _, weapon := range rGetPlayerWeapons.GetWeapons() {
			if weapon.GetName() == weaponMsg {
				c.Payload.WeaponID = weapon.GetID()
				return false
			}
		}

		return true
	// ##################################################################################################
	// Verifico Conferma
	// ##################################################################################################
	case 2:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetCrafterRepairController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Chiedo al player di indicare quale arma vuole riparare
	// ##################################################################################################
	case 0:
		var weaponsKeyboard [][]tgbotapi.KeyboardButton

		// Recupero armi player
		var rGetPlayerWeapons *pb.GetPlayerWeaponsResponse
		rGetPlayerWeapons, _ = config.App.Server.Connection.GetPlayerWeapons(helpers.NewContext(1), &pb.GetPlayerWeaponsRequest{
			PlayerID: c.Player.ID,
		})

		for _, weapon := range rGetPlayerWeapons.GetWeapons() {
			// Mostro solo armi danneggiate
			if weapon.Durability < weapon.DurabilityCap {
				weaponsKeyboard = append(weaponsKeyboard, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						fmt.Sprintf("%s - 🔧%v/%v",
							weapon.GetName(),
							weapon.GetDurability(),
							weapon.GetDurabilityCap(),
						),
					),
				))
			}
		}

		weaponsKeyboard = append(weaponsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.repair.what"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       weaponsKeyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 1
	// ##################################################################################################
	// Chiedo Conferma al player
	// ##################################################################################################
	case 1:
		// Recupero dettagli arma
		var rGetWeaponByID *pb.GetWeaponByIDResponse
		if rGetWeaponByID, err = config.App.Server.Connection.GetWeaponByID(helpers.NewContext(1), &pb.GetWeaponByIDRequest{
			ID: c.Payload.WeaponID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero informazioni riparazione
		var rCrafterGetRepairWeaponInfo *pb.CrafterGetRepairWeaponInfoResponse
		if rCrafterGetRepairWeaponInfo, err = config.App.Server.Connection.CrafterGetRepairWeaponInfo(helpers.NewContext(1), &pb.CrafterGetRepairWeaponInfoRequest{
			WeaponID: c.Payload.WeaponID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.repair.confirm",
			rGetWeaponByID.GetWeapon().GetName(),
			rCrafterGetRepairWeaponInfo.GetValue(),
		))

		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Stage = 2
	// ##################################################################################################
	// Riparo arma
	// ##################################################################################################
	case 2:
		// Avvio riparazione arma
		_, err = config.App.Server.Connection.CrafterRepairWeapon(helpers.NewContext(1), &pb.CrafterRepairWeaponRequest{
			PlayerID: c.Player.ID,
			WeaponID: c.Payload.WeaponID,
		})

		if err != nil && strings.Contains(err.Error(), "player dont have enough money") {
			// Potrebbero esserci stati degli errori come per esempio la mancanza di monete
			errorMsg := helpers.NewMessage(c.ChatID,
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

		msg := helpers.NewMessage(c.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.crafting.repair.completed"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Completed = true
	}
}
