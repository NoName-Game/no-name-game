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
// SafePlanetMarketAuctionsBuyController
// ====================================
type SafePlanetMarketAuctionsBuyController struct {
	Payload struct {
		ItemType  string
		AuctionID uint32
		Bid       int32
	}
	Controller
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMarketAuctionsBuyController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
	if !c.InitController(Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.market.auctions.buy",
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
				4: {"route.breaker.menu"},
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
func (c *SafePlanetMarketAuctionsBuyController) Validator() (hasErrors bool) {
	var err error

	switch c.CurrentState.Stage {
	case 1:
		// Verifico qualche categoria di aste vuole vedere
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "armor") {
			c.Payload.ItemType = "armors"
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "weapon") {
			c.Payload.ItemType = "weapons"
		}

	// ##################################################################################################
	// Recupero informazioni riguardo l'asta selezionata
	// ##################################################################################################
	case 2:
		auctionSplit := strings.Split(c.Update.Message.Text, "#")
		if len(auctionSplit)-1 > 0 {
			var auctionID int
			if auctionID, err = strconv.Atoi(auctionSplit[1]); err != nil {
				return true
			}

			c.Payload.AuctionID = uint32(auctionID)
		}

	// ##################################################################################################
	// Verifico l'offerta fatta
	// ##################################################################################################
	case 3:
		// TODO: Verifico se il player possiede in banca il totale
		// TODO: Verifico se l'asta è ancora attiva

		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_100") {
			c.Payload.Bid = 100
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_250") {
			c.Payload.Bid = 250
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_500") {
			c.Payload.Bid = 500
		} else {
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
func (c *SafePlanetMarketAuctionsBuyController) Stage() {
	var err error
	switch c.CurrentState.Stage {
	// ##################################################################################################
	// Faccio scegliere la categoria
	// ##################################################################################################
	case 0:
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.which_category"))
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

		// Recupero aste per la categoria scelta
		switch c.Payload.ItemType {
		case "armors":
			var rGetAllAuctionsByCategory *pb.GetAllAuctionsByCategoryResponse
			if rGetAllAuctionsByCategory, err = config.App.Server.Connection.GetAllAuctionsByCategory(helpers.NewContext(1), &pb.GetAllAuctionsByCategoryRequest{
				ItemCategory: 0,
			}); err != nil {
				c.Logger.Panic(err)
			}

			if len(rGetAllAuctionsByCategory.GetAuctions()) > 0 {
				// Ciclo aste
				for _, auction := range rGetAllAuctionsByCategory.GetAuctions() {

					// Recupero dettagli arma
					var rGetArmorByID *pb.GetArmorByIDResponse
					if rGetArmorByID, err = config.App.Server.Connection.GetArmorByID(helpers.NewContext(1), &pb.GetArmorByIDRequest{
						ArmorID: auction.GetItemID(),
					}); err != nil {
						c.Logger.Panic(err)
					}

					keyboardRow := tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							fmt.Sprintf(
								"%s (%s) 🛡 #%v",
								// helpers.Trans(c.Player.Language.Slug, "equip"),
								rGetArmorByID.GetArmor().GetName(),
								rGetArmorByID.GetArmor().GetRarity().GetSlug(),
								auction.GetID(),
							),
						),
					)

					keyboardRows = append(keyboardRows, keyboardRow)
				}
			}

		case "weapons":
			var rGetAllAuctionsByCategory *pb.GetAllAuctionsByCategoryResponse
			if rGetAllAuctionsByCategory, err = config.App.Server.Connection.GetAllAuctionsByCategory(helpers.NewContext(1), &pb.GetAllAuctionsByCategoryRequest{
				ItemCategory: 1,
			}); err != nil {
				c.Logger.Panic(err)
			}

			if len(rGetAllAuctionsByCategory.GetAuctions()) > 0 {
				// Ciclo aste
				for _, auction := range rGetAllAuctionsByCategory.GetAuctions() {
					// Recupero dettagli arma
					var rGetWeaponByID *pb.GetWeaponByIDResponse
					if rGetWeaponByID, err = config.App.Server.Connection.GetWeaponByID(helpers.NewContext(1), &pb.GetWeaponByIDRequest{
						ID: auction.GetItemID(),
					}); err != nil {
						c.Logger.Panic(err)
					}

					keyboardRow := tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(
							fmt.Sprintf(
								"%s (%s) -🩸%v #%v",
								rGetWeaponByID.GetWeapon().GetName(),
								rGetWeaponByID.GetWeapon().GetRarity().GetSlug(),
								math.Round(rGetWeaponByID.GetWeapon().GetRawDamage()),
								auction.GetID(),
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
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.which_item"))
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
		// Recupero dettagli asta
		var rGetAuctionByID *pb.GetAuctionByIDResponse
		if rGetAuctionByID, err = config.App.Server.Connection.GetAuctionByID(helpers.NewContext(1), &pb.GetAuctionByIDRequest{
			AuctionID: c.Payload.AuctionID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero dettagli offerte
		var rGetAuctionBids *pb.GetAuctionBidsResponse
		if rGetAuctionBids, err = config.App.Server.Connection.GetAuctionBids(helpers.NewContext(1), &pb.GetAuctionBidsRequest{
			AuctionID: c.Payload.AuctionID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero dettagli arma
		var itemDetails string
		switch rGetAuctionByID.GetAuction().GetItemCategory() {
		case pb.AuctionItemCategoryEnum_ARMOR:
			// Recupero dettagli armatura
			var rGetArmorByID *pb.GetArmorByIDResponse
			if rGetArmorByID, err = config.App.Server.Connection.GetArmorByID(helpers.NewContext(1), &pb.GetArmorByIDRequest{
				ArmorID: rGetAuctionByID.GetAuction().GetItemID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			itemDetails = fmt.Sprintf(
				"\n<b>(%s)</b> (%s) - [%v, %v%%, %v%%] 🎖%v",
				rGetArmorByID.GetArmor().Name,
				strings.ToUpper(rGetArmorByID.GetArmor().Rarity.Slug),
				math.Round(rGetArmorByID.GetArmor().Defense),
				math.Round(rGetArmorByID.GetArmor().Evasion),
				math.Round(rGetArmorByID.GetArmor().Halving),
				rGetArmorByID.GetArmor().Rarity.LevelToEuip,
			)

		case pb.AuctionItemCategoryEnum_WEAPON:
			// Recupero dettagli arma
			var rGetWeaponByID *pb.GetWeaponByIDResponse
			if rGetWeaponByID, err = config.App.Server.Connection.GetWeaponByID(helpers.NewContext(1), &pb.GetWeaponByIDRequest{
				ID: rGetAuctionByID.GetAuction().GetItemID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			itemDetails = fmt.Sprintf(
				"<b>(%s)</b> (%s) - [%v, %v%%, %v] 🎖%v",
				rGetWeaponByID.GetWeapon().Name,
				strings.ToUpper(rGetWeaponByID.GetWeapon().Rarity.Slug),
				math.Round(rGetWeaponByID.GetWeapon().RawDamage),
				math.Round(rGetWeaponByID.GetWeapon().Precision),
				rGetWeaponByID.GetWeapon().Durability,
				rGetWeaponByID.GetWeapon().Rarity.LevelToEuip,
			)
		}

		// Chiedo al player di inserire il prezzo minimo di partenza
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.auction_details",
			rGetAuctionByID.GetAuction().GetPlayer().GetUsername(),
			itemDetails,
			rGetAuctionByID.GetAuction().GetMinPrice(),
			rGetAuctionBids.GetTotalBid(), rGetAuctionBids.GetLastBid().GetPlayer().GetUsername(),
		))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_100")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_250")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_500")),
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
		c.CurrentState.Stage = 3
	case 3:
		// Recupero dettagli offerte asta
		var rGetAuctionBids *pb.GetAuctionBidsResponse
		if rGetAuctionBids, err = config.App.Server.Connection.GetAuctionBids(helpers.NewContext(1), &pb.GetAuctionBidsRequest{
			AuctionID: c.Payload.AuctionID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Calcolo totale
		finalBid := rGetAuctionBids.GetTotalBid() + c.Payload.Bid

		// Chiedo conferma offerta
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.confirm_bid",
			c.Payload.Bid,
			finalBid,
		))

		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "confirm")),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
			),
		)

		// Invio
		if _, err = helpers.SendMessage(msg); err != nil {
			c.Logger.Panic(err)
		}

		// Aggiorno stato
		c.CurrentState.Stage = 4

	case 4:
		// Chiamo ws per registrare
		if _, err = config.App.Server.Connection.NewAuctionBid(helpers.NewContext(1), &pb.NewAuctionBidRequest{
			AuctionID: c.Payload.AuctionID,
			PlayerID:  c.Player.ID,
			Bid:       c.Payload.Bid,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Mando ok creazine asta
		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_ok"))
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

		c.CurrentState.Stage = 2
		c.Stage()
	}

}
