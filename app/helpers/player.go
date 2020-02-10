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
		services.SendMessage(msg)

		err = errors.New("missing username")
		return player, err
	}

	// Verifico se esiste già un player registrato
	player, err = providers.FindPlayerByUsername(user.UserName)
	if err != nil {
		return player, err
	}

	// Se il player non esiste allora lo registro
	if player.ID <= 0 {
		// Recupero lingua di default
		var language nnsdk.Language
		language, err = providers.FindLanguageBySlug("it")
		if err != nil {
			return player, err
		}

		// Registro player
		player, err = providers.SignIn(nnsdk.Player{
			Username:   user.UserName,
			ChatID:     int64(user.ID),
			LanguageID: language.ID,
		})

		return player, err
	}

	return
}

// GetPlayerStateByFunction - Check if function exist in player states
func GetPlayerStateByFunction(player nnsdk.Player, function string) (playerState nnsdk.PlayerState, err error) {
	for i, state := range player.States {
		if state.Function == function {
			playerState = player.States[i]
			return playerState, nil
		}
	}

	return playerState, errors.New("State not found!")
}

// CheckPlayerHaveOneEquippedWeapon
// Verifica se il player ha almeno un'arma equipaggiata
func CheckPlayerHaveOneEquippedWeapon(player nnsdk.Player) bool {
	for _, weapon := range player.Weapons {
		if *weapon.Equipped == true {
			return true
		}
	}

	return false
}
