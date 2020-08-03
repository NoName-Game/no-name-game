package helpers

import (
	"context"
	"errors"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := services.NnSDK.FindPlayerByUsername(ctx, &pb.FindPlayerByUsernameRequest{
		Username: user.UserName,
	})
	if err != nil {
		return player, err
	}

	player = response.GetPlayer()

	// Se il player non esiste allora lo registro
	if player.ID == 0 {
		// Recupero lingua di default

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		response, err := services.NnSDK.FindLanguageBySlug(ctx, &pb.FindLanguageBySlugRequest{
			Slug: "it",
		})
		if err != nil {
			return player, err
		}

		var language *pb.Language
		language = response.GetLanguage()

		// Registro player
		ctxSignIn, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		responseSignIn, err := services.NnSDK.SignIn(ctxSignIn, &pb.SignInRequest{
			Player: &pb.Player{
				Username:   user.UserName,
				ChatID:     int64(user.ID),
				LanguageID: language.ID,
			},
		})
		if err != nil {
			return player, err
		}

		player = responseSignIn.GetPlayer()

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
	for _, weapon := range player.GetWeapons() {
		if weapon.GetEquipped() {
			return true
		}
	}

	return false
}

// GetPlayerCurrentPlanet
// Recupera il pianeta corrente del player
func GetPlayerCurrentPlanet(player *pb.Player) (planet *pb.Planet, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	responseLastPosition, err := services.NnSDK.GetPlayerLastPosition(ctx, &pb.GetPlayerLastPositionRequest{
		PlayerID: player.GetID(),
	})
	if err != nil {
		return planet, err
	}

	// Recupero ultima posizione del player, dando per scontato che sia
	// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
	var lastPosition *pb.PlayerPosition
	lastPosition = responseLastPosition.GetPlayerPosition()

	// Dalla ultima posizione recupero il pianeta corrente
	responsePlanet, err := services.NnSDK.GetPlanetByCoordinate(ctx, &pb.GetPlanetByCoordinateRequest{
		X: lastPosition.GetX(),
		Y: lastPosition.GetY(),
		Z: lastPosition.GetZ(),
	})
	if err != nil {
		return planet, err
	}

	return responsePlanet.GetPlanet(), nil
}
