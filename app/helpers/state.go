package helpers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/services"
)

// GetRedisState - Metodo generico per il recupero degli stati di un player
func GetRedisState(player nnsdk.Player) (controller string, err error) {
	controller, err = services.Redis.Get(strconv.FormatUint(uint64(player.ID), 10)).Result()
	return
}

// SetRedisState - Metodo generico per il settaggio di uno stato su redis di un determinato player
func SetRedisState(player nnsdk.Player, data string) (err error) {
	err = services.Redis.Set(strconv.FormatUint(uint64(player.ID), 10), data, 0).Err()
	return
}

// DelRedisState - Metodo generico per la cancellazione degli stati di un determinato player
func DelRedisState(player nnsdk.Player) (err error) {
	err = services.Redis.Del(strconv.FormatUint(uint64(player.ID), 10)).Err()
	return
}

// GetHuntingRedisState - get hunting state in Redis
func GetHuntingRedisState(IDMap uint, player nnsdk.Player) (huntingMap nnsdk.Map) {
	state, err := services.Redis.Get(fmt.Sprintf("hunting_%v_%v", IDMap, player.ID)).Result()
	if err != nil {
		services.ErrorHandler("Error getting hunting state in redis", err)
	}

	json.Unmarshal([]byte(state), &huntingMap)
	return
}

// SetRedisState - set function state in Redis
func SetHuntingRedisState(IDMap uint, player nnsdk.Player, value interface{}) {
	jsonValue, _ := json.Marshal(value)
	err := services.Redis.Set(fmt.Sprintf("hunting_%v_%v", IDMap, player.ID), string(jsonValue), 0).Err()
	if err != nil {
		services.ErrorHandler("Error SET player state in redis", err)
	}
}

// CheckState - Verifica ed effettua controlli sullo stato del player in un determinato controller
func CheckState(player nnsdk.Player, controller string, payload interface{}, father uint) (playerState nnsdk.PlayerState, isNewState bool, err error) {
	// Filtro gli stati del player recuperando lo stato appartente a questa specifica rotta
	playerState, _ = GetPlayerStateByFunction(player.States, controller)

	// Non ho trovato nessuna corrispondenza creo una nuova
	if playerState.ID <= 0 {
		jsonPayload, _ := json.Marshal(payload)

		// Creo il nuovo stato
		playerState, err = providers.CreatePlayerState(nnsdk.PlayerState{
			Controller: controller,
			PlayerID:   player.ID,
			Payload:    string(jsonPayload),
			Father:     father,
		})

		// Ritoro errore se non riesco a creare lo stato
		if err != nil {
			return playerState, true, err
		}

		// Ritorno indicando che si tratta di un nuovo stato, questo
		// mi servirà per escludere il validator
		isNewState = true
	}

	// Se tutto ok imposto e setto il nuovo stato su redis
	err = SetRedisState(player, controller)
	if err != nil {
		return playerState, false, err
	}

	return
}

// FinishAndCompleteState - finish and set completed in playerstate
// func FinishAndCompleteState(state nnsdk.PlayerState, player nnsdk.Player) (err error) {
// 	// Stupid poninter stupid json pff
// 	t := new(bool)
// 	*t = true
// 	state.Completed = t
//
// 	// Aggiunro
// 	state, err = providers.UpdatePlayerState(state) // Update
// 	if err != nil {
// 		return err
// 	}
//
// 	state, err = providers.DeletePlayerState(state) // Delete
// 	if err != nil {
// 		return err
// 	}
//
// 	err = DelRedisState(player)
// 	return
// }

// DeleteRedisAndDbState - delete redis and db state
func DeleteRedisAndDbState(player nnsdk.Player) {
	// rediState, _ := GetRedisState(player)
	//
	// if rediState != "" {
	// 	playerState, _ := GetPlayerStateByFunction(player, rediState)
	// 	_, err := providers.DeletePlayerState(playerState) // Delete
	// 	if err != nil {
	// 		// FIXME: loggare ma non entrare in panic
	// 		// services.ErrorHandler("Error delete player state", err)
	// 	}
	// }
	//
	// DelRedisState(player)
}

// UnmarshalPayload - Unmarshal payload state
func UnmarshalPayload(payload string, funcInterface interface{}) {
	if payload != "" {
		err := json.Unmarshal([]byte(payload), &funcInterface)
		if err != nil {
			services.ErrorHandler("Error unmarshal payload", err)
		}
	}
}
