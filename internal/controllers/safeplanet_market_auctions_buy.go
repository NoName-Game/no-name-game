package controllers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"nn-grpc/build/pb"
	"nn-telegram/config"
	"nn-telegram/internal/helpers"
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

func (c *SafePlanetMarketAuctionsBuyController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
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
				1: {"route.breaker.menu", "route.breaker.back"},
				2: {"route.breaker.menu", "route.breaker.clears", "route.breaker.back"},
				3: {"route.breaker.menu", "route.breaker.clears", "route.breaker.back"},
				4: {"route.breaker.menu", "route.breaker.back"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMarketAuctionsBuyController) Handle(player *pb.Player, update tgbotapi.Update) {
	// Verifico se è impossibile inizializzare
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
func (c *SafePlanetMarketAuctionsBuyController) Validator() (hasErrors bool) {
	var err error

	switch c.CurrentState.Stage {
	case 1:
		category := strings.Split(c.Update.Message.Text, " (")

		// Verifico qualche categoria di aste vuole vedere
		if category[0] == helpers.Trans(c.Player.Language.Slug, "armor") {
			c.Payload.ItemType = "armors"
		} else if category[0] == helpers.Trans(c.Player.Language.Slug, "weapon") {
			c.Payload.ItemType = "weapons"
		} else {
			return true
		}

	// ##################################################################################################
	// Recupero informazioni riguardo l'asta selezionata
	// ##################################################################################################
	case 2:
		if strings.Contains(c.Update.Message.Text, "#") {
			auctionSplit := strings.Split(c.Update.Message.Text, "#")
			if len(auctionSplit)-1 > 0 {
				var auctionID int
				if auctionID, err = strconv.Atoi(auctionSplit[1]); err != nil {
					return true
				}

				c.Payload.AuctionID = uint32(auctionID)
				return false
			}
		}

		c.Validation.Message = helpers.Trans(c.Player.GetLanguage().GetSlug(), "safeplanet.market.auctions.buy.select_one_auction")
		return true

	// ##################################################################################################
	// Verifico l'offerta fatta
	// ##################################################################################################
	case 3:
		if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_100") {
			c.Payload.Bid = 100
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_250") {
			c.Payload.Bid = 250
		} else if c.Update.Message.Text == helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_500") {
			c.Payload.Bid = 500
		} else {
			return true
		}

		// Recupero dettagli asta
		var rGetAuctionByID *pb.GetAuctionByIDResponse
		if rGetAuctionByID, err = config.App.Server.Connection.GetAuctionByID(helpers.NewContext(1), &pb.GetAuctionByIDRequest{
			AuctionID: c.Payload.AuctionID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero dettagli offerte asta
		var rGetAuctionBids *pb.GetAuctionBidsResponse
		if rGetAuctionBids, err = config.App.Server.Connection.GetAuctionBids(helpers.NewContext(1), &pb.GetAuctionBidsRequest{
			AuctionID: c.Payload.AuctionID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Verifico che il player non sia l'owner dell'asta
		if rGetAuctionByID.GetAuction().GetPlayerID() == c.Player.GetID() {
			c.Validation.Message = helpers.Trans(c.Player.GetLanguage().GetSlug(), "safeplanet.market.auctions.buy.error_owner")
			return true
		}

		// Recupero budget player, ovvero i soldi che possiede in banca
		var rGetPlayerEconomy *pb.GetPlayerEconomyResponse
		if rGetPlayerEconomy, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
			PlayerID:    c.Player.GetID(),
			EconomyType: pb.GetPlayerEconomyRequest_BANK,
		}); err != nil {
			c.Logger.Panic(err)
		}

		totalBid := rGetAuctionBids.GetTotalBid()
		currentBid := c.Payload.Bid

		// Verifico se il player possiede in banca il totale
		if rGetPlayerEconomy.GetValue() < totalBid+currentBid {
			c.Validation.Message = helpers.Trans(c.Player.GetLanguage().GetSlug(), "safeplanet.market.auctions.buy.budget_low")
			return true
		}

		// Verifico se l'asta è aperta
		var closeAt time.Time
		if closeAt, err = helpers.GetEndTime(rGetAuctionByID.GetAuction().GetCloseAt(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		if time.Now().After(closeAt) {
			c.Validation.Message = helpers.Trans(c.Player.GetLanguage().GetSlug(), "safeplanet.market.auctions.buy.auction_closed")
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
		// Recupero quante aste ci sono per la categoria armature
		var rGetAllArmorAuction *pb.GetAllAuctionsByCategoryResponse
		if rGetAllArmorAuction, err = config.App.Server.Connection.GetAllAuctionsByCategory(helpers.NewContext(1), &pb.GetAllAuctionsByCategoryRequest{
			ItemCategory: 0,
			PlayerID:     c.Player.GetID(),
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero quante aste ci sono per la categoria armi
		var rGetAllWeaponAuctions *pb.GetAllAuctionsByCategoryResponse
		if rGetAllWeaponAuctions, err = config.App.Server.Connection.GetAllAuctionsByCategory(helpers.NewContext(1), &pb.GetAllAuctionsByCategoryRequest{
			PlayerID:     c.Player.GetID(),
			ItemCategory: 1,
		}); err != nil {
			c.Logger.Panic(err)
		}

		msg := helpers.NewMessage(c.Player.ChatID, helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.which_category"))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					fmt.Sprintf("%s (%v)", helpers.Trans(c.Player.Language.Slug, "armor"), len(rGetAllArmorAuction.GetAuctions())),
				),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(
					fmt.Sprintf("%s (%v)", helpers.Trans(c.Player.Language.Slug, "weapon"), len(rGetAllWeaponAuctions.GetAuctions())),
				),
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
			// Verifico se ci sono delle aste al quale il player ha fatto un'offerta
			var rGetAllPlayerOfferAuctionsByCategory *pb.GetAllPlayerOfferAuctionsByCategoryResponse
			if rGetAllPlayerOfferAuctionsByCategory, err = config.App.Server.Connection.GetAllPlayerOfferAuctionsByCategory(helpers.NewContext(1), &pb.GetAllPlayerOfferAuctionsByCategoryRequest{
				PlayerID:     c.Player.ID,
				ItemCategory: 0,
			}); err != nil {
				c.Logger.Panic(err)
			}

			if len(rGetAllPlayerOfferAuctionsByCategory.GetAuctions()) > 0 {
				keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.GetLanguage().GetSlug(), "safeplanet.market.auctions.banner_offer_auction")),
				))

				// Ciclo aste tutte le aste
				for _, auction := range rGetAllPlayerOfferAuctionsByCategory.GetAuctions() {
					// Recupero dettagli arma
					keyboardRow, _ := helpers.AuctionWeaponKeyboard(auction.GetItemID(), auction.GetID())
					keyboardRows = append(keyboardRows, keyboardRow)
				}
			}

			// Ciclo tutte le aste
			var rGetAllAuctionsByCategory *pb.GetAllAuctionsByCategoryResponse
			if rGetAllAuctionsByCategory, err = config.App.Server.Connection.GetAllAuctionsByCategory(helpers.NewContext(1), &pb.GetAllAuctionsByCategoryRequest{
				ItemCategory: 0,
				PlayerID:     c.Player.GetID(),
			}); err != nil {
				c.Logger.Panic(err)
			}

			if len(rGetAllAuctionsByCategory.GetAuctions()) > 0 {
				keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.GetLanguage().GetSlug(), "safeplanet.market.auctions.banner_all_auction")),
				))

				// Ciclo tutte le aste
				for _, auction := range rGetAllAuctionsByCategory.GetAuctions() {
					keyboardRow, _ := helpers.AuctionArmorKeyboard(auction.GetItemID(), auction.GetID())
					keyboardRows = append(keyboardRows, keyboardRow)
				}
			}

		case "weapons":
			// Verifico se ci sono delle aste al quale il player ha fatto un'offerta
			var rGetAllPlayerOfferAuctionsByCategory *pb.GetAllPlayerOfferAuctionsByCategoryResponse
			if rGetAllPlayerOfferAuctionsByCategory, err = config.App.Server.Connection.GetAllPlayerOfferAuctionsByCategory(helpers.NewContext(1), &pb.GetAllPlayerOfferAuctionsByCategoryRequest{
				PlayerID:     c.Player.ID,
				ItemCategory: 1,
			}); err != nil {
				c.Logger.Panic(err)
			}

			if len(rGetAllPlayerOfferAuctionsByCategory.GetAuctions()) > 0 {
				keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.GetLanguage().GetSlug(), "safeplanet.market.auctions.banner_offer_auction")),
				))

				// Ciclo aste tutte le aste
				for _, auction := range rGetAllPlayerOfferAuctionsByCategory.GetAuctions() {
					// Recupero dettagli arma
					keyboardRow, _ := helpers.AuctionWeaponKeyboard(auction.GetItemID(), auction.GetID())
					keyboardRows = append(keyboardRows, keyboardRow)
				}
			}

			// Ciclo tutte le aste per questa categoria
			var rGetAllAuctionsByCategory *pb.GetAllAuctionsByCategoryResponse
			if rGetAllAuctionsByCategory, err = config.App.Server.Connection.GetAllAuctionsByCategory(helpers.NewContext(1), &pb.GetAllAuctionsByCategoryRequest{
				PlayerID:     c.Player.GetID(),
				ItemCategory: 1,
			}); err != nil {
				c.Logger.Panic(err)
			}

			if len(rGetAllAuctionsByCategory.GetAuctions()) > 0 {
				keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.GetLanguage().GetSlug(), "safeplanet.market.auctions.banner_all_auction")),
				))

				// Ciclo aste tutte le aste
				for _, auction := range rGetAllAuctionsByCategory.GetAuctions() {
					// Recupero dettagli arma
					keyboardRow, _ := helpers.AuctionWeaponKeyboard(auction.GetItemID(), auction.GetID())
					keyboardRows = append(keyboardRows, keyboardRow)
				}
			}
		}

		keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
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

		// Costruisco dettagli item
		var itemDetails string
		if itemDetails, err = helpers.AuctionItemFormatter(c.Payload.AuctionID); err != nil {
			c.Logger.Panic(err)
		}

		var bidKeyboard [][]tgbotapi.KeyboardButton
		// Verifico se l'asta è ancora aperta
		var closeAt time.Time
		if closeAt, err = helpers.GetEndTime(rGetAuctionByID.GetAuction().GetCloseAt(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		if time.Now().Before(closeAt) {
			bidKeyboard = append(bidKeyboard,
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_100")),
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_250")),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.bid_500")),
				),
			)
		}

		bidKeyboard = append(bidKeyboard, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(helpers.Trans(c.Player.Language.Slug, "route.breaker.back")),
		))

		auctionDetails := helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.auction_details",
			rGetAuctionByID.GetAuction().GetPlayer().GetUsername(),
			closeAt.Format("15:04:05 02/01"),
			itemDetails,
			rGetAuctionByID.GetAuction().GetMinPrice(),
		)

		if rGetAuctionBids.GetLastBid() != nil {
			auctionDetails += helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.last_offer",
				rGetAuctionBids.GetTotalBid(), rGetAuctionBids.GetLastBid().GetPlayer().GetUsername(),
			)
		}

		// Recupero budget player, ovvero i soldi che possiede in banca
		var rGetPlayerEconomy *pb.GetPlayerEconomyResponse
		if rGetPlayerEconomy, err = config.App.Server.Connection.GetPlayerEconomy(helpers.NewContext(1), &pb.GetPlayerEconomyRequest{
			PlayerID:    c.Player.GetID(),
			EconomyType: pb.GetPlayerEconomyRequest_BANK,
		}); err != nil {
			c.Logger.Panic(err)
		}

		playerBudget := helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.buy.player_budget", rGetPlayerEconomy.GetValue())

		// Chiedo al player di inserire il prezzo minimo di partenza
		msg := helpers.NewMessage(c.Player.ChatID, fmt.Sprintf("%s\n%s", auctionDetails, playerBudget))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard:       bidKeyboard,
		}

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
		// Recupero ultimo offerente
		var rGetAuctionBids *pb.GetAuctionBidsResponse
		rGetAuctionBids, _ = config.App.Server.Connection.GetAuctionBids(helpers.NewContext(1), &pb.GetAuctionBidsRequest{
			AuctionID: c.Payload.AuctionID,
		})

		if rGetAuctionBids.GetLastBid() != nil {
			auctionItemDetails, _ := helpers.AuctionItemFormatter(c.Payload.AuctionID)

			// Invio messaggio all'ultimo offerente indicandogli che la sua offerta è stata superata
			msgToOldOffer := helpers.NewMessage(rGetAuctionBids.GetLastBid().GetPlayer().GetChatID(),
				helpers.Trans(c.Player.Language.Slug, "notification.auction.offer_higher", auctionItemDetails),
			)
			msgToOldOffer.ParseMode = tgbotapi.ModeHTML
			if _, err = helpers.SendMessage(msgToOldOffer); err != nil {
				c.Logger.Panic(err)
			}
		}

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
