package helpers

import (
	"fmt"
	"math"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
)

func AuctionItemFormatter(auctionID uint32) (itemDetails string, err error) {
	// Recupero dettagli asta
	var rGetAuctionByID *pb.GetAuctionByIDResponse
	if rGetAuctionByID, err = config.App.Server.Connection.GetAuctionByID(NewContext(1), &pb.GetAuctionByIDRequest{
		AuctionID: auctionID,
	}); err != nil {
		return itemDetails, err
	}

	// Recupero dettagli arma
	switch rGetAuctionByID.GetAuction().GetItemCategory() {
	case pb.AuctionItemCategoryEnum_ARMOR:
		// Recupero dettagli armatura
		var rGetArmorByID *pb.GetArmorByIDResponse
		if rGetArmorByID, err = config.App.Server.Connection.GetArmorByID(NewContext(1), &pb.GetArmorByIDRequest{
			ArmorID: rGetAuctionByID.GetAuction().GetItemID(),
		}); err != nil {
			return itemDetails, err
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
		if rGetWeaponByID, err = config.App.Server.Connection.GetWeaponByID(NewContext(1), &pb.GetWeaponByIDRequest{
			ID: rGetAuctionByID.GetAuction().GetItemID(),
		}); err != nil {
			return itemDetails, err
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

	return
}

func AuctionArmorKeyboard(amorID uint32, auctionID uint32) (keyboardRow []tgbotapi.KeyboardButton, err error) {
	// Recupero dettagli arma
	var rGetArmorByID *pb.GetArmorByIDResponse
	if rGetArmorByID, err = config.App.Server.Connection.GetArmorByID(NewContext(1), &pb.GetArmorByIDRequest{
		ArmorID: amorID,
	}); err != nil {
		return keyboardRow, err
	}

	return tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(
			fmt.Sprintf(
				"%s (%s) 🛡 #%v",
				// helpers.Trans(c.Player.Language.Slug, "equip"),
				rGetArmorByID.GetArmor().GetName(),
				rGetArmorByID.GetArmor().GetRarity().GetSlug(),
				auctionID,
			),
		),
	), nil
}

func AuctionWeaponKeyboard(weaponID uint32, auctionID uint32) (keyboardRow []tgbotapi.KeyboardButton, err error) {
	var rGetWeaponByID *pb.GetWeaponByIDResponse
	if rGetWeaponByID, err = config.App.Server.Connection.GetWeaponByID(NewContext(1), &pb.GetWeaponByIDRequest{
		ID: weaponID,
	}); err != nil {
		return keyboardRow, err
	}

	return tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(
			fmt.Sprintf(
				"%s (%s) -🩸%v #%v",
				rGetWeaponByID.GetWeapon().GetName(),
				rGetWeaponByID.GetWeapon().GetRarity().GetSlug(),
				math.Round(rGetWeaponByID.GetWeapon().GetRawDamage()),
				auctionID,
			),
		),
	), nil
}
