package controllers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetHangarShipsController
// ====================================
type SafePlanetHangarShipsController struct {
	Controller
	Payload struct {
		CategoryID uint32
		ShipID     uint32
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetHangarShipsController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.hangar.ships",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetHangarController{},
				FromStage: 1,
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
func (c *SafePlanetHangarShipsController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico categoria passata
	// ##################################################################################################
	case 1:
		var err error
		var rGetAllShipCategories *pb.GetAllShipCategoriesResponse
		if rGetAllShipCategories, err = config.App.Server.Connection.GetAllShipCategories(helpers.NewContext(1), &pb.GetAllShipCategoriesRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		for _, category := range rGetAllShipCategories.GetShipCategories() {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ship.category.%s", category.GetSlug())) == c.Update.Message.Text {
				c.Payload.CategoryID = category.GetID()
				return false
			}
		}

		return true

	// ##################################################################################################
	// Verifico nave passata
	// ##################################################################################################
	case 2:
		var err error
		var rGetPlayerShips *pb.GetPlayerShipsResponse
		if rGetPlayerShips, err = config.App.Server.Connection.GetPlayerShips(helpers.NewContext(1), &pb.GetPlayerShipsRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		for _, ship := range rGetPlayerShips.GetShips() {
			if c.Update.Message.Text == ship.GetName() {
				c.Payload.ShipID = ship.GetID()
				return false
			}
		}

		return true

	// ##################################################################################################
	// Verifico conferma sostituzione nave
	// ##################################################################################################
	case 3:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.ships.confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetHangarShipsController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Recupero categorie della nave
	// ##################################################################################################
	case 0:
		var rGetAllShipCategories *pb.GetAllShipCategoriesResponse
		if rGetAllShipCategories, err = config.App.Server.Connection.GetAllShipCategories(helpers.NewContext(1), &pb.GetAllShipCategoriesRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		var categoriesKeyboard [][]tgbotapi.KeyboardButton
		for _, category := range rGetAllShipCategories.GetShipCategories() {
			categoriesKeyboard = append(categoriesKeyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ship.category.%s", category.GetSlug())),
				),
			))
		}

		categoriesKeyboard = append(categoriesKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.create.intro"))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       categoriesKeyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	// ##################################################################################################
	// Recupero navi per la categoria scelta
	// ##################################################################################################
	case 1:
		var rGetPlayerShips *pb.GetPlayerShipsResponse
		if rGetPlayerShips, err = config.App.Server.Connection.GetPlayerShips(helpers.NewContext(1), &pb.GetPlayerShipsRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		var shipsKeyboard [][]tgbotapi.KeyboardButton
		for _, ship := range rGetPlayerShips.GetShips() {
			if ship.GetShipCategory().GetID() == c.Payload.CategoryID {
				shipsKeyboard = append(shipsKeyboard, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						ship.GetName(),
					),
				))
			}
		}

		shipsKeyboard = append(shipsKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.ships.list"))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       shipsKeyboard,
		}
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2

	// ##################################################################################################
	// Recupero dettagli nave
	// ##################################################################################################
	case 2:
		var rGetShipByID *pb.GetShipByIDResponse
		if rGetShipByID, err = config.App.Server.Connection.GetShipByID(helpers.NewContext(1), &pb.GetShipByIDRequest{
			ShipID: c.Payload.ShipID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		var text string
		text = fmt.Sprintf(" Ecco i dettagli della nave: %s\n",
			rGetShipByID.GetShip().GetName(),
		)

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, text)
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.ships.confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3

	// ##################################################################################################
	// Equipaggia nave
	// ##################################################################################################
	case 3:
		if _, err = config.App.Server.Connection.EquipShip(helpers.NewContext(1), &pb.EquipShipRequest{
			PlayerID: c.Player.ID,
			ShipID:   c.Payload.ShipID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.ships.equip_ok"))
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
			),
		)

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Completo lo stato
		c.CurrentState.Completed = true
	}

	return
}
