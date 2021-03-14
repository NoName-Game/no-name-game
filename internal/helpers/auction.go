package helpers

import (
	"fmt"
	"math"

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

		itemDetails = ArmorFormatter(rGetArmorByID.GetArmor())

	case pb.AuctionItemCategoryEnum_WEAPON:
		// Recupero dettagli arma
		var rGetWeaponByID *pb.GetWeaponByIDResponse
		if rGetWeaponByID, err = config.App.Server.Connection.GetWeaponByID(NewContext(1), &pb.GetWeaponByIDRequest{
			ID: rGetAuctionByID.GetAuction().GetItemID(),
		}); err != nil {
			return itemDetails, err
		}

		itemDetails = WeaponFormatter(rGetWeaponByID.GetWeapon())
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

// Verifica se l'armatura si trova in un'asta
func CheckIfArmorInAuction(playerID uint32, armorID uint32) (inAuction bool, err error) {
	// Recupero quante aste ci sono per la categoria armi
	var rGetAllArmorAuction *pb.GetAllAuctionsByCategoryResponse
	if rGetAllArmorAuction, err = config.App.Server.Connection.GetAllAuctionsByCategory(NewContext(1), &pb.GetAllAuctionsByCategoryRequest{
		ItemCategory: 0,
		PlayerID:     0, // Non filtro per nessun utente
	}); err != nil {
		return
	}

	for _, auction := range rGetAllArmorAuction.GetAuctions() {
		if auction.GetItemID() == armorID {
			return true, nil
		}
	}

	return false, nil
}

// Verifica se l'arma si trova in un'asta
func CheckIfWeaponInAuction(playerID uint32, weaponID uint32) (inAuction bool, err error) {
	// Recupero quante aste ci sono per la categoria armi
	var rGetAllWeaponAuctions *pb.GetAllAuctionsByCategoryResponse
	if rGetAllWeaponAuctions, err = config.App.Server.Connection.GetAllAuctionsByCategory(NewContext(1), &pb.GetAllAuctionsByCategoryRequest{
		PlayerID:     0, // Non filtro per nessun utente
		ItemCategory: 1,
	}); err != nil {
		return
	}

	for _, auction := range rGetAllWeaponAuctions.GetAuctions() {
		if auction.GetItemID() == weaponID {
			return true, nil
		}
	}

	return false, nil
}
