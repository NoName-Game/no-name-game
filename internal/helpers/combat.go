package helpers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	// Hunting Fight Actions
	FightStart = InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "start_fight",
	}

	FightUp = InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "up",
	}

	FightDown = InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "down",
	}

	FightReturnMap = InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "return_map",
	}

	FightHit = InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "hit",
	}

	FightNoAction = InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "no_action",
	}

	FightPlayerDie = InlineDataStruct{
		C:  "hunting",
		AT: "fight",
		A:  "player_die",
	}
)

// PlayerFightKeyboard
func PlayerFightKeyboard(player *pb.Player, baseFightKeyboard [][]tgbotapi.InlineKeyboardButton) (keyboard *tgbotapi.InlineKeyboardMarkup, err error) {
	newfightKeyboard := new(tgbotapi.InlineKeyboardMarkup)

	// #######################
	// Usabili: recupero quali item possono essere usati in combattimento
	// #######################
	// Ciclo pozioni per ID item
	for _, itemID := range []uint32{1, 2, 3} {
		var rGetItemByID *pb.GetItemByIDResponse
		if rGetItemByID, err = config.App.Server.Connection.GetItemByID(NewContext(1), &pb.GetItemByIDRequest{
			ItemID: itemID,
		}); err != nil {
			return nil, err
		}

		var rGetPlayerItemByID *pb.GetPlayerItemByIDResponse
		if rGetPlayerItemByID, err = config.App.Server.Connection.GetPlayerItemByID(NewContext(1), &pb.GetPlayerItemByIDRequest{
			PlayerID: player.GetID(),
			ItemID:   itemID,
		}); err != nil {
			return nil, err
		}

		// Aggiunto tasto solo se la quantità del player è > 0
		if rGetPlayerItemByID.GetPlayerInventory().GetQuantity() > 0 {
			var potionStruct = InlineDataStruct{C: "hunting", AT: "fight", A: "use", D: rGetItemByID.GetItem().GetID()}
			newfightKeyboard.InlineKeyboard = append(newfightKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("%s (%v)",
						Trans(player.Language.Slug, fmt.Sprintf("items.%s", rGetItemByID.GetItem().GetSlug())),
						rGetPlayerItemByID.GetPlayerInventory().GetQuantity(),
					),
					potionStruct.GetDataString(),
				),
			))
		}
	}

	// #######################
	// Keyboard Selezione, attacco e fuga
	// #######################
	newfightKeyboard.InlineKeyboard = append(newfightKeyboard.InlineKeyboard, baseFightKeyboard...)

	// #######################
	// Abilità
	// #######################
	// Ciclo le abilità al combattimento
	for _, abilityID := range []uint32{7, 8} {
		// Verifico se il player possiede abilità di comattimento o difesa
		var rCheckIfPlayerHaveAbility *pb.CheckIfPlayerHaveAbilityResponse
		if rCheckIfPlayerHaveAbility, err = config.App.Server.Connection.CheckIfPlayerHaveAbility(NewContext(1), &pb.CheckIfPlayerHaveAbilityRequest{
			PlayerID:  player.GetID(),
			AbilityID: abilityID, // Attacco pesante
		}); err != nil {
			return nil, err
		}

		if rCheckIfPlayerHaveAbility.GetHaveAbility() {
			// Appendo abilità player
			var dataAbilityStruct = InlineDataStruct{C: "hunting", AT: "fight", A: "hit", SA: "ability", D: rCheckIfPlayerHaveAbility.GetAbility().GetID()}
			newfightKeyboard.InlineKeyboard = append(newfightKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					Trans(player.Language.Slug, fmt.Sprintf("safeplanet.accademy.ability.%s", rCheckIfPlayerHaveAbility.GetAbility().GetSlug())),
					dataAbilityStruct.GetDataString(),
				),
			))
		}
	}

	return newfightKeyboard, nil
}

// UseItem
func UseItem(player *pb.Player, itemID uint32, MessageID int) (err error) {
	// Recupero dettagli item che si vuole usare
	var rGetItemByID *pb.GetItemByIDResponse
	if rGetItemByID, err = config.App.Server.Connection.GetItemByID(NewContext(1), &pb.GetItemByIDRequest{
		ItemID: itemID,
	}); err != nil {
		return err
	}

	// Richiamo il ws per usare l'item selezionato
	if _, err = config.App.Server.Connection.UseItem(NewContext(1), &pb.UseItemRequest{
		PlayerID: player.GetID(),
		ItemID:   rGetItemByID.GetItem().GetID(),
	}); err != nil {
		return err
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(Trans(player.GetLanguage().GetSlug(), "continue"), FightNoAction.GetDataString()),
		),
	)

	var combactMessage tgbotapi.EditMessageTextConfig
	combactMessage = NewEditMessage(player.ChatID, MessageID,
		Trans(player.Language.Slug, "combat.use_item",
			Trans(player.GetLanguage().GetSlug(), fmt.Sprintf("items.%s", rGetItemByID.GetItem().GetSlug())),
			rGetItemByID.GetItem().GetValue(),
		),
	)
	combactMessage.ReplyMarkup = &keyboard
	combactMessage.ParseMode = tgbotapi.ModeHTML
	if _, err = SendMessage(combactMessage); err != nil {
		return err
	}

	return
}
