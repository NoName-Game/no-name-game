package controllers

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// SafePlanetMarketAuctionsSellController
// ====================================
type SafePlanetMarketAuctionsSellController struct {
	Payload struct {
		ItemType string
		ItemID   uint32
		Price    int
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMarketAuctionsSellController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.market.auctions.sell",
			Payload:    &c.Payload,
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetMarketAuctionsController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
				1: {"route.breaker.menu"},
				2: {"route.breaker.menu"},
				3: {"route.breaker.menu", "route.breaker.clears"},
				4: {"route.breaker.menu", "route.breaker.clears"},
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
func (c *SafePlanetMarketAuctionsSellController) Validator() (hasErrors bool) {
	var err error

	// TODO: verifico che il player non abbia più di tre aste contemporanee

	switch c.CurrentState.Stage {
	case 1:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "armor") {
			c.Payload.ItemType = "armors"
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "weapon") {
			c.Payload.ItemType = "weapons"
		}

	// ##################################################################################################
	// Verifico se il player possiete l'item passato
	// ##################################################################################################
	case 2:
		// Verifico se il player possiede la risorsa passata
		var haveResource bool

		// Recupero nome item che il player vuole usare
		equipmentName := strings.Split(c.Update.Message.Text, " (")[0]

		switch c.Payload.ItemType {
		case "armors":
			var rGetArmorByName *pb.GetArmorByPlayerAndNameResponse
			if rGetArmorByName, err = config.App.Server.Connection.GetArmorByPlayerAndName(helpers.NewContext(1), &pb.GetArmorByPlayerAndNameRequest{
				Name:     equipmentName,
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Verifico se è l'armatura già presente in un'asta
			var inAuction bool
			if inAuction, err = helpers.CheckIfArmorInAuction(c.Player.ID, rGetArmorByName.GetArmor().GetID()); err != nil {
				c.Logger.Panic(err)
			}

			if inAuction {
				c.Validation.Message = helpers.Trans(c.Player.GetLanguage().GetSlug(), "safeplanet.market.auctions.sell.item_already_in_auction")
				return true
			}

			// Verifico se appartiene correttamente al player
			if rGetArmorByName.GetArmor().GetPlayerID() == c.Player.ID {
				c.Payload.ItemID = rGetArmorByName.GetArmor().GetID()
				haveResource = true
			}

			haveResource = false
		case "weapons":
			var rGetWeaponByName *pb.GetWeaponByPlayerAndNameResponse
			if rGetWeaponByName, err = config.App.Server.Connection.GetWeaponByPlayerAndName(helpers.NewContext(1), &pb.GetWeaponByPlayerAndNameRequest{
				Name:     equipmentName,
				PlayerID: c.Player.ID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			// Verifico se è un'arma già presente in un'asta
			var inAuction bool
			if inAuction, err = helpers.CheckIfWeaponInAuction(c.Player.ID, rGetWeaponByName.GetWeapon().GetID()); err != nil {
				c.Logger.Panic(err)
			}

			if inAuction {
				c.Validation.Message = helpers.Trans(c.Player.GetLanguage().GetSlug(), "safeplanet.market.auctions.sell.item_already_in_auction")
				return true
			}

			// Verifico se appartiene correttamente al player
			if rGetWeaponByName.GetWeapon().GetPlayerID() == c.Player.ID {
				c.Payload.ItemID = rGetWeaponByName.GetWeapon().GetID()
				haveResource = true
			}
		}

		if !haveResource {
			return true
		}

	// ##################################################################################################
	// Verifico il prezzo minimo insertio dall'utente
	// ##################################################################################################
	case 3:
		// TODO: verificare prezzo
		if c.Payload.Price, err = strconv.Atoi(c.Update.Message.Text); err != nil {
			return true
		}
	case 4:
		if c.Update.Message.Text != helpers.Trans(c.Player.Language.Slug, "confirm") {
			return true
		}
	}

	return false
}

// ====================================
// Stage
// ====================================
func (c *SafePlanetMarketAuctionsSellController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Faccio scegliere la categoria
	// ##################################################################################################
	case 0:
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.sell.which_category"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "armor")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "weapon")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		// Invio messaggio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 1

	case 1:
		// Costruisco keyboard
		var keyboardRows [][]tgbotapi.KeyboardButton

		// Recupero item per la categoria scelta
		switch c.Payload.ItemType {
		case "armors":
			// Recupero armature non equipaggiate filtrate per la categoria scelta
			var rGetPlayerArmors *pb.GetPlayerArmorsResponse
			if rGetPlayerArmors, err = config.App.Server.Connection.GetPlayerArmors(helpers.NewContext(1), &pb.GetPlayerArmorsRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			if len(rGetPlayerArmors.GetArmors()) > 0 {
				// Ciclo armature del player
				for _, armor := range rGetPlayerArmors.GetArmors() {
					// Posso mettere all'asta solo le armature non equipaggiate
					if armor.GetEquipped() {
						continue
					}

					keyboardRow := tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							fmt.Sprintf(
								"%s (%s) 🛡 🎖%v",
								// helpers.Trans(c.Player.Language.Slug, "equip"),
								armor.Name,
								strings.ToUpper(armor.Rarity.Slug),
								armor.Rarity.LevelToEuip,
							),
						),
					)

					keyboardRows = append(keyboardRows, keyboardRow)
				}
			}

		case "weapons":
			// Recupero nuovamente armi player, richiamando la rotta dedicata
			// in questa maniera posso filtrare per quelle che non sono equipaggiate
			var rGetPlayerWeapons *pb.GetPlayerWeaponsResponse
			if rGetPlayerWeapons, err = config.App.Server.Connection.GetPlayerWeapons(helpers.NewContext(1), &pb.GetPlayerWeaponsRequest{
				PlayerID: c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			if len(rGetPlayerWeapons.GetWeapons()) > 0 {
				// Ciclo armi player
				for _, weapon := range rGetPlayerWeapons.GetWeapons() {
					// Posso mettere all'asta solo le armi non equipaggiate
					if weapon.GetEquipped() {
						continue
					}

					keyboardRow := tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							fmt.Sprintf(
								"%s (%s) -🩸%v 🎖%v",
								weapon.Name,
								strings.ToUpper(weapon.Rarity.Slug),
								math.Round(weapon.RawDamage),
								weapon.Rarity.LevelToEuip,
							),
						),
					)
					keyboardRows = append(keyboardRows, keyboardRow)
				}
			}
		}

		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
		))

		// Mando messaggio
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.sell.which_item"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       keyboardRows,
		}

		// Invio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 2
	case 2:
		// Chiedo al player di inserire il prezzo minimo di partenza
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.sell.min_price"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		// Invio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 3
	case 3:
		// Reucpero dettagli item
		var itemName string
		switch c.Payload.ItemType {
		case "armors":
			var rGetPlayerArmors *pb.GetArmorByIDResponse
			if rGetPlayerArmors, err = config.App.Server.Connection.GetArmorByID(helpers.NewContext(1), &pb.GetArmorByIDRequest{
				ArmorID: c.Payload.ItemID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			itemName = helpers.ArmorFormatter(rGetPlayerArmors.GetArmor())
		case "weapons":
			var rGetPlayerWeapons *pb.GetWeaponByIDResponse
			if rGetPlayerWeapons, err = config.App.Server.Connection.GetWeaponByID(helpers.NewContext(1), &pb.GetWeaponByIDRequest{
				ID: c.Payload.ItemID,
			}); err != nil {
				c.Logger.Panic(err)
			}

			itemName = helpers.WeaponFormatter(rGetPlayerWeapons.GetWeapon())
		}

		// Chiedo conferma al player
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug,
			"safeplanet.market.auctions.sell.confirm",
			itemName,
			c.Payload.Price,
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

		// Invio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 4

	case 4:
		var itemCategory pb.AuctionItemCategoryEnum
		switch c.Payload.ItemType {
		case "amors":
			itemCategory = pb.AuctionItemCategoryEnum_ARMOR
		case "weapons":
			itemCategory = pb.AuctionItemCategoryEnum_WEAPON
		}

		// Creo nuova asta
		if _, err = config.App.Server.Connection.NewAuction(helpers.NewContext(1), &pb.NewAuctionRequest{
			PlayerID:     c.Player.ID,
			ItemID:       c.Payload.ItemID,
			ItemCategory: itemCategory,
			MinPrice:     int32(c.Payload.Price),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Mando ok creazine asta
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.sell.ok"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.menu")),
			),
		)

		// Invio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		c.CurrentState.Completed = true
	}

}
