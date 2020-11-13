package controllers

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"

	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetHangarCreateController
// ====================================
type SafePlanetHangarCreateController struct {
	Controller
	Payload struct {
		CategoryID          uint32
		RarityID            uint32
		CompleteWithDiamond bool
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetHangarCreateController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.hangar.create",
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
func (c *SafePlanetHangarCreateController) Validator() (hasErrors bool) {
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Verifico se è stata passata una categoria di nave corretta
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
	// Verifico quale rarità vuole il player
	// ##################################################################################################
	case 2:
		rarityMsg := strings.Split(c.Update.Message.Text, " -")[0]

		var err error
		var rGetAllCraftableRarities *pb.GetAllCraftableRaritiesResponse
		if rGetAllCraftableRarities, err = config.App.Server.Connection.GetAllCraftableRarities(helpers.NewContext(1), &pb.GetAllCraftableRaritiesRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		for _, rarity := range rGetAllCraftableRarities.GetRarities() {
			if helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("rarity.%s", rarity.GetSlug())) == rarityMsg {
				c.Payload.RarityID = rarity.GetID()
				return false
			}
		}

		return true

	// ##################################################################################################
	// Verifico conferma acquisto
	// ##################################################################################################
	case 3:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	// ##################################################################################################
	// Verifico completamento costruzione nave
	// ##################################################################################################
	case 4:
		var err error
		var rCheckCreateShip *pb.CheckCreateShipResponse
		if rCheckCreateShip, err = config.App.Server.Connection.CheckCreateShip(helpers.NewContext(1), &pb.CheckCreateShipRequest{
			PlayerID: c.Player.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Il crafter sta già portando a terminre un lavoro per questo player
		if !rCheckCreateShip.GetFinishShipCreating() {
			if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "complete_with_diamond") {
				c.Payload.CompleteWithDiamond = true
				return false
			}

			var finishAt time.Time
			if finishAt, err = helpers.GetEndTime(rCheckCreateShip.GetShipCreatingEndTime(), c.Player); err != nil {
				c.Logger.Panic(err)
			}

			// Calcolo diamanti del player
			var rGetPlayerEconomyDiamond *pb.GetPlayerEconomyResponse
			if rGetPlayerEconomyDiamond, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
				PlayerID:    c.Player.GetID(),
				EconomyType: pb.GetPlayerEconomyRequest_DIAMOND,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Invio messaggio recap fine viaggio
			c.Validation.Message = helpers.Trans(
				c.Player.Language.Slug,
				"safeplanet.hangar.create.wait",
				finishAt.Format("15:04:05"),
				rGetPlayerEconomyDiamond.GetValue(),
			)

			// Aggiungi possibilità di velocizzare
			c.Validation.ReplyKeyboard = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "complete_with_diamond"),
					),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(
						helpers.Trans(c.Player.Language.Slug, "route.breaker.more"),
					),
				),
			)

			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetHangarCreateController) Stage() {
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
		msg.ParseMode = tgbotapi.ModeMarkdown
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
	// In questo stage chiedo di quale rarità
	// ##################################################################################################
	case 1:
		var rGetAllCraftableRarities *pb.GetAllCraftableRaritiesResponse
		if rGetAllCraftableRarities, err = config.App.Server.Connection.GetAllCraftableRarities(helpers.NewContext(1), &pb.GetAllCraftableRaritiesRequest{}); err != nil {
			c.Logger.Panic(err)
		}

		var raritiesKeyboard [][]tgbotapi.KeyboardButton
		for _, rarity := range rGetAllCraftableRarities.GetRarities() {
			// Recupero informazioni costruzione
			var rGetCreateShipInfo *pb.GetCreateShipInfoResponse
			if rGetCreateShipInfo, err = config.App.Server.Connection.GetCreateShipInfo(helpers.NewContext(1), &pb.GetCreateShipInfoRequest{
				RarityID:   rarity.ID,
				CategoryID: c.Payload.CategoryID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			raritiesKeyboard = append(raritiesKeyboard, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					fmt.Sprintf("%s - 💰%v",
						helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("rarity.%s", rarity.GetSlug())),
						rGetCreateShipInfo.GetPrice(),
					),
				),
			))
		}

		raritiesKeyboard = append(raritiesKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.create.chose_rarity"))
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       raritiesKeyboard,
		}

		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2

	// ##################################################################################################
	// Chiedo conferma
	// ##################################################################################################
	case 2:
		// Recupero informazioni costruzione
		var rGetCreateShipInfo *pb.GetCreateShipInfoResponse
		if rGetCreateShipInfo, err = config.App.Server.Connection.GetCreateShipInfo(helpers.NewContext(1), &pb.GetCreateShipInfoRequest{
			RarityID:   c.Payload.RarityID,
			CategoryID: c.Payload.CategoryID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero dettagli rarità richiesta
		var rGetRarityByID *pb.GetRarityByIDResponse
		if rGetRarityByID, err = config.App.Server.Connection.GetRarityByID(helpers.NewContext(1), &pb.GetRarityByIDRequest{
			RarityID: c.Payload.RarityID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero dettagli categoria richiesta
		var rGetShipCategoryByID *pb.GetShipCategoryByIDResponse
		if rGetShipCategoryByID, err = config.App.Server.Connection.GetShipCategoryByID(helpers.NewContext(1), &pb.GetShipCategoryByIDRequest{
			ShipCategoryID: c.Payload.CategoryID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.create.confirm",
			helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("ship.category.%s", rGetShipCategoryByID.GetShipCategory().GetSlug())), // rarity
			helpers.Trans(c.Player.Language.Slug, fmt.Sprintf("rarity.%s", rGetRarityByID.GetRarity().GetSlug())),                    // rarity
			rGetCreateShipInfo.GetPrice(),
		))
		msg.ParseMode = tgbotapi.ModeMarkdown
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
		c.CurrentState.Stage = 3

	// ##################################################################################################
	// Avvio costruzione nave
	// ##################################################################################################
	case 3:
		var err error
		var rStartCreateShip *pb.StartCreateShipResponse
		rStartCreateShip, err = config.App.Server.Connection.StartCreateShip(helpers.NewContext(1), &pb.StartCreateShipRequest{
			RarityID:   c.Payload.RarityID,
			CategoryID: c.Payload.CategoryID,
			PlayerID:   c.Player.ID,
		})

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

		// Recupero orario fine lavorazione
		var finishAt time.Time
		if finishAt, err = helpers.GetEndTime(rStartCreateShip.GetShipCreateEndTime(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID,
			helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.create.start", finishAt.Format("15:04:05 01/02")),
		)
		msg.ParseMode = tgbotapi.ModeMarkdown
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 4
		c.ForceBackTo = true

	// ##################################################################################################
	// Fine costruzione nave
	// ##################################################################################################
	case 4:
		var ship *pb.Ship
		// Verifico se ha gemmato
		if c.Payload.CompleteWithDiamond {
			var rEndCreateShipDiamond *pb.EndCreateShipResponse
			if rEndCreateShipDiamond, err = config.App.Server.Connection.EndCreateShipDiamond(helpers.NewContext(1), &pb.EndCreateShipRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				// Messaggio errore completamento
				msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.create.complete_diamond_error"))
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.more")),
					),
				)

				if _, err = helpers.SendMessage(msg); err != nil {
					c.Logger.Panic(err)
				}

				// Fondamentale, esco senza chiudere
				c.ForceBackTo = true
				return
			}

			ship = rEndCreateShipDiamond.GetShip()
		} else {
			// Costruisco chiamata per aggiornare posizione e scalare il quantitativo di carburante usato
			var rEndCreateShip *pb.EndCreateShipResponse
			if rEndCreateShip, err = config.App.Server.Connection.EndCreateShip(helpers.NewContext(1), &pb.EndCreateShipRequest{
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			ship = rEndCreateShip.GetShip()
		}

		// Invio messaggio
		msg := helpers.NewMessage(c.Update.Message.Chat.ID, helpers.Trans(c.Player.Language.Slug, "safeplanet.hangar.create.end", ship.GetName()))
		msg.ParseMode = tgbotapi.ModeMarkdown
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
