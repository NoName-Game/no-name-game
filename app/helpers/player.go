package helpers

import (
	"errors"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
	"bitbucket.org/no-name-game/nn-telegram/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// HandleUser - Eseguo varie verifiche per controllare il player
func HandleUser(update tgbotapi.Update) (player nnsdk.Player, err error) {
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
	var playerProvider providers.PlayerProvider
	player, err = playerProvider.FindPlayerByUsername(user.UserName)
	if err != nil {
		return player, err
	}

	// Se il player non esiste allora lo registro
	if player.ID == 0 {
		// Recupero lingua di default
		var language nnsdk.Language
		var languageProvider providers.LanguageProvider
		language, err = languageProvider.FindLanguageBySlug("it")
		if err != nil {
			return player, err
		}

		// Registro player
		player, err = playerProvider.SignIn(nnsdk.Player{
			Username:   user.UserName,
			ChatID:     int64(user.ID),
			LanguageID: language.ID,
		})

		return player, err
	}

	return
}

// GetPlayerStateByFunction - Check if function exist in player states
func GetPlayerStateByFunction(states nnsdk.PlayerStates, controller string) (playerState nnsdk.PlayerState, err error) {
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
func CheckPlayerHaveOneEquippedWeapon(player nnsdk.Player) bool {
	for _, weapon := range player.Weapons {
		if *weapon.Equipped {
			return true
		}
	}

	return false
}

// GetPlayerCurrentPlanet
// Recupera il pianeta corrente del player
func GetPlayerCurrentPlanet(player nnsdk.Player) (planet nnsdk.Planet, err error) {
	var playerProvider providers.PlayerProvider
	var planetProvider providers.PlanetProvider

	// Recupero ultima posizione del player, dando per scontato che sia
	// la posizione del pianeta e quindi della mappa corrente che si vuole recuperare
	var lastPosition nnsdk.PlayerPosition
	lastPosition, err = playerProvider.GetPlayerLastPosition(player)
	if err != nil {
		return planet, err
	}

	// Dalla ultima posizione recupero il pianeta corrente
	planet, err = planetProvider.GetPlanetByCoordinate(lastPosition.X, lastPosition.Y, lastPosition.Z)
	if err != nil {
		return planet, err
	}

	return planet, nil
}
