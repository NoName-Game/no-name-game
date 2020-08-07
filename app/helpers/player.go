package helpers

import (
	"errors"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/services"

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
		msg := services.NewMessage(update.Message.Chat.ID, Trans("en", "miss_username"))
		_, err = services.SendMessage(msg)
		if err != nil {
			return player, err
		}

		err = errors.New("missing username")
		return player, err
	}

	// Verifico se esiste già un player registrato
	rGetPlayerByUsername, _ := services.NnSDK.GetPlayerByUsername(NewContext(1), &pb.GetPlayerByUsernameRequest{
		Username: user.UserName,
	})

	// Recupero player
	player = rGetPlayerByUsername.GetPlayer()

	// Se il player non esiste allora lo registro
	if player.GetID() == 0 {
		// Recupero lingua di default
		rFindLanguageBySlug, err := services.NnSDK.FindLanguageBySlug(NewContext(1), &pb.FindLanguageBySlugRequest{
			Slug: "it",
		})
		if err != nil {
			return player, err
		}

		// Registro player
		rSignIn, err := services.NnSDK.SignIn(NewContext(10), &pb.SignInRequest{
			Username:   user.UserName,
			ChatID:     int64(user.ID), // TODO: !? Non dovrebbe esser chatID !?
			LanguageID: rFindLanguageBySlug.GetLanguage().GetID(),
		})
		if err != nil {
			return player, err
		}

		player = rSignIn.GetPlayer()

		return player, err
	}

	return
}

// GetPlayerStateByFunction - Check if function exist in player states
func GetPlayerStateByFunction(states []*pb.PlayerState, controller string) (playerState *pb.PlayerState, err error) {
	for i, state := range states {
		if state.Controller == controller {
			playerState = states[i]
			return playerState, nil
		}
	}

	err = errors.New("state not found")
	return playerState, err
}

// CheckPlayerHaveOneEquippedWeapon
// Verifica se il player ha almeno un'arma equipaggiata
func CheckPlayerHaveOneEquippedWeapon(player *pb.Player) bool {
	rGetPlayerWeapons, _ := services.NnSDK.GetPlayerWeaponEquipped(NewContext(1), &pb.GetPlayerWeaponEquippedRequest{
		PlayerID: player.GetID(),
	})

	if rGetPlayerWeapons.GetWeapon() != nil && rGetPlayerWeapons.GetWeapon().GetID() > 0 {
		return true
	}

	return false
}

// GetPlayerCurrentPlanet
// Recupera il pianeta corrente del player
func GetPlayerCurrentPlanet(player *pb.Player) (planet *pb.Planet, err error) {
	// Recupero ultima posizione del player, dando per scontato che sia
	// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
	rGetPlayerLastPosition, err := services.NnSDK.GetPlayerLastPosition(NewContext(1), &pb.GetPlayerLastPositionRequest{
		PlayerID: player.GetID(),
	})
	if err != nil {
		return planet, err
	}

	// Dalla ultima posizione recupero il pianeta corrente
	responsePlanet, err := services.NnSDK.GetPlanetByCoordinate(NewContext(1), &pb.GetPlanetByCoordinateRequest{
		X: rGetPlayerLastPosition.GetPlayerPosition().GetX(),
		Y: rGetPlayerLastPosition.GetPlayerPosition().GetY(),
		Z: rGetPlayerLastPosition.GetPlayerPosition().GetZ(),
	})
	if err != nil {
		return planet, err
	}

	return responsePlanet.GetPlanet(), nil
}
