package helpers

import (
	"errors"
	"github.com/sirupsen/logrus"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
	"bitbucket.org/no-name-game/nn-telegram/config"

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
		// Registro player
		var rSignIn *pb.SignInResponse
		if rSignIn, err = config.App.Server.Connection.SignIn(NewContext(10), &pb.SignInRequest{
			Username: user.UserName,
			ChatID:   int64(user.ID), // TODO: !? Non dovrebbe esser chatID !?
		}); err != nil {
			return player, err
		}

		player = rSignIn.GetPlayer()

		return player, err
	}

	return
}

// CheckPlayerHaveActiveActivity - Verifica se il player ha in corso una determinata attività
func CheckPlayerHaveActiveActivity(activities []*pb.PlayerActivity, controller string) bool {
	for _, state := range activities {
		if state.Controller == controller {
			return true
		}
	}

	return false
}

// GetPlayerPosition - Recupera l'ultima posizione del player
func GetPlayerPosition(playerID uint32) (position *pb.Planet, err error) {
	// Tento di recuperare posizione da cache
	if position, err = GetPlayerPlanetPositionInCache(playerID); err != nil {
		// Recupero ultima posizione nota del player
		var rGetPlayerCurrentPlanet *pb.GetPlayerCurrentPlanetResponse
		if rGetPlayerCurrentPlanet, err = config.App.Server.Connection.GetPlayerCurrentPlanet(NewContext(1), &pb.GetPlayerCurrentPlanetRequest{
			PlayerID: playerID,
		}); err != nil {
			return nil, err
		}

		// Verifico se il player si trova su un pianeta valido
		if rGetPlayerCurrentPlanet.GetPlanet() == nil {
			return nil, err
		}

		position = rGetPlayerCurrentPlanet.GetPlanet()

		// Creo cache posizione
		if err = SetPlayerPlanetPositionInCache(playerID, position); err != nil {
			logrus.Errorf("error creating player position cache: %s", err.Error())
		}
	}

	return
}

func SortPlayerArmor(armors []*pb.Armor) []*pb.Armor {
	result := make([]*pb.Armor, 4)
	// testa, braccia, petto, gambe
	for _, armor := range armors {
		switch armor.GetArmorCategory().GetSlug() {
		case "helmet":
			result[0] = armor
		case "glove":
			result[1] = armor
		case "chest":
			result[2] = armor
		case "boots":
			result[3] = armor
		}
	}
	return result
}
