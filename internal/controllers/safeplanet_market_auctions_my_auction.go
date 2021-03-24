package controllers

import (
	"time"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	"bitbucket.org/no-name-game/nn-telegram/internal/helpers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ====================================
// Market
// ====================================
type SafePlanetMarketAuctionsMyAuctionController struct {
	Controller
}

func (c *SafePlanetMarketAuctionsMyAuctionController) Configuration(player *pb.Player, update tgbotapi.Update) Controller {
	return Controller{
		Player: player,
		Update: update,
		CurrentState: ControllerCurrentState{
			Controller: "route.safeplanet.market.auctions.my_auction",
		},
		Configurations: ControllerConfigurations{
			ControllerBack: ControllerBack{
				To:        &SafePlanetMarketController{},
				FromStage: 0,
			},
			PlanetType: []string{"safe"},
			BreakerPerStage: map[int32][]string{
				0: {"route.breaker.menu"},
			},
		},
	}
}

// ====================================
// Handle
// ====================================
func (c *SafePlanetMarketAuctionsMyAuctionController) Handle(player *pb.Player, update tgbotapi.Update) {
	var err error

	// Init Controller
	if !c.InitController(c.Configuration(player, update)) {
		return
	}

	// Recupero aste del player
	var rGetAllPlayerAuctions *pb.GetAllPlayerAuctionsResponse
	if rGetAllPlayerAuctions, err = config.App.Server.Connection.GetAllPlayerAuctions(helpers.NewContext(1), &pb.GetAllPlayerAuctionsRequest{
		PlayerID: c.Player.ID,
	}); err != nil {
		c.Logger.Panic(err)
	}

	var recapPlayerAuctions string
	for _, auction := range rGetAllPlayerAuctions.GetAuctions() {
		// Costruisco dettagli item
		var itemDetails string
		if itemDetails, err = helpers.AuctionItemFormatter(auction.ID); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero dettagli offerte
		var rGetAuctionBids *pb.GetAuctionBidsResponse
		if rGetAuctionBids, err = config.App.Server.Connection.GetAuctionBids(helpers.NewContext(1), &pb.GetAuctionBidsRequest{
			AuctionID: auction.ID,
		}); err != nil {
			c.Logger.Panic(err)
		}

		// Recupero data fine asta
		var closeAt time.Time
		if closeAt, err = helpers.GetEndTime(auction.GetCloseAt(), c.Player); err != nil {
			c.Logger.Panic(err)
		}

		recapPlayerAuctions += helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.auction_details",
			auction.GetPlayer().GetUsername(),
			closeAt.Format("15:04:05 02/01"),
			itemDetails,
			auction.GetMinPrice(),
		)

		if rGetAuctionBids.GetLastBid() != nil {
			recapPlayerAuctions += helpers.Trans(c.Player.Language.Slug, "safeplanet.market.auctions.last_offer",
				rGetAuctionBids.GetTotalBid(), rGetAuctionBids.GetLastBid().GetPlayer().GetUsername(),
			)
		}

		recapPlayerAuctions += "\n---------------------------------------\n"
	}

	var recapMessage = helpers.Trans(player.Language.Slug, "safeplanet.market.auctions.my_auction.no_auction")
	if recapPlayerAuctions != "" {
		recapMessage = helpers.Trans(player.Language.Slug, "safeplanet.market.auctions.my_auction.info_recap", recapPlayerAuctions)
	}

	msg := helpers.NewMessage(c.ChatID, recapMessage)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err = helpers.SendMessage(msg); err != nil {
		c.Logger.Panic(err)
	}
}

func (c *SafePlanetMarketAuctionsMyAuctionController) Validator() bool {
	return false
}

func (c *SafePlanetMarketAuctionsMyAuctionController) Stage() {
	//
}
