package helpers

import (
	"errors"

	"bitbucket.org/no-name-game/nn-telegram/config"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// HandleUser - Eseguo varie verifiche per controllare il player
func HandleUser(update tgbotapi.Update) (player *pb.Player, err error) {
	// Recupero utente filtrandolo per tipologia di messaggio
	var user *tgbotapi.User
	if update.Message != nil {
		// Se è un messaggio normale
		user = update.Message.From
	} else if update.CallbackQuery != nil {
		// Se è una callback di un messaggio con action inline
		user = update.CallbackQuery.From
	} else {
		err = errors.New("unsupported type of message")
		return player, err
	}

	// Controllo se il player non ha un username
	if user.UserName == "" {
		// Mando un messaggio dicendogli di inserire un username
		msg := NewMessage(update.Message.Chat.ID, Trans("en", "miss_username"))
		if _, err = SendMessage(msg); err != nil {
			return player, err
		}

		err = errors.New("missing username")
		return player, err
	}

	// Verifico se esiste già un player registrato
	rGetPlayerByUsername, _ := config.App.Server.Connection.GetPlayerByUsername(NewContext(1), &pb.GetPlayerByUsernameRequest{
		Username: user.UserName,
	})

	// Recupero player
	player = rGetPlayerByUsername.GetPlayer()

	// Se il player non esiste allora lo registro
	if player.GetID() == 0 {
		// Recupero lingua di default
		var rGetLanguageBySlug *pb.GetLanguageBySlugResponse
		if rGetLanguageBySlug, err = config.App.Server.Connection.GetLanguageBySlug(NewContext(1), &pb.GetLanguageBySlugRequest{
			Slug: "it",
		}); err != nil {
			return player, err
		}

		// Registro player
		var rSignIn *pb.SignInResponse
		if rSignIn, err = config.App.Server.Connection.SignIn(NewContext(10), &pb.SignInRequest{
			Username:   user.UserName,
			ChatID:     int64(user.ID), // TODO: !? Non dovrebbe esser chatID !?
			LanguageID: rGetLanguageBySlug.GetLanguage().GetID(),
		}); err != nil {
			return player, err
		}

		player = rSignIn.GetPlayer()

		return player, err
	}

	return
}

// CheckPlayerHaveOneEquippedWeapon
// Verifica se il player ha almeno un'arma equipaggiata
func CheckPlayerHaveOneEquippedWeapon(player *pb.Player) bool {
	rGetPlayerWeapons, _ := config.App.Server.Connection.GetPlayerWeaponEquipped(NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
		PlayerID: player.GetID(),
	})

	if rGetPlayerWeapons.GetWeapon() != nil && rGetPlayerWeapons.GetWeapon().GetID() > 0 {
		return true
	}

	return false
}
