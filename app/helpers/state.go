package helpers

import (
	"encoding/json"
	"fmt"
	"strconv"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/services"
)

// GetRedisState - Metodo generico per il recupero degli stati di un player
func GetRedisState(player pb.Player) (controller string, err error) {
	controller, err = services.Redis.Get(strconv.FormatUint(uint64(player.GetID()), 10)).Result()
	return
}

// SetRedisState - Metodo generico per il settaggio di uno stato su redis di un determinato player
func SetRedisState(player pb.Player, data string) (err error) {
	err = services.Redis.Set(strconv.FormatUint(uint64(player.GetID()), 10), data, 0).Err()
	return
}

// DelRedisState - Metodo generico per la cancellazione degli stati di un determinato player
func DelRedisState(player pb.Player) (err error) {
	err = services.Redis.Del(strconv.FormatUint(uint64(player.GetID()), 10)).Err()
	return
}

// DeleteRedisAndDbState - Cancella record da redis e dal DB
// func DeleteRedisAndDbState(player nnsdk.Player) (err error) {
// 	rediState, _ := GetRedisState(player)
//
// 	if rediState != "" {
// 		var playerState nnsdk.PlayerState
// 		playerState, err = GetPlayerStateByFunction(player.States, rediState)
// 		if err != nil {
// 			return err
// 		}
//
// 		var playerStateProvider providers.PlayerStateProvider
// 		_, err = playerStateProvider.DeletePlayerState(playerState) // Delete
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	err = DelRedisState(player)
// 	return err
// }

// GetRedisPlayerHuntingPosition - recupero posizione di una player in una specifica mappa
func GetRedisPlayerHuntingPosition(maps *pb.Maps, player *pb.Player, positionType string) (value int32, err error) {
	var state string
	state, err = services.Redis.Get(fmt.Sprintf("hunting_map_%v_player_%v_position_%s", maps.ID, player.ID, positionType)).Result()
	if err != nil {
		return value, err
	}

	conversion, err := strconv.Atoi(state)
	if err != nil {
		return value, err
	}

	value = int32(conversion)

	return value, err
}

// SetRedisPlayerHuntingPosition - Imposto posizione di un player su una determinata mappa
func SetRedisPlayerHuntingPosition(maps *pb.Maps, player *pb.Player, positionType string, value int32) (err error) {
	err = services.Redis.Set(fmt.Sprintf("hunting_map_%v_player_%v_position_%s", maps.ID, player.ID, positionType), value, 0).Err()
	return
}

// GetRedisMapHunting - Recupera mappa su redis
func GetRedisMapHunting(IDMap uint32) (maps *pb.Maps, err error) {
	var state string
	state, err = services.Redis.Get(fmt.Sprintf("hunting_map_%v", IDMap)).Result()
	if err != nil {
		return maps, err
	}

	err = json.Unmarshal([]byte(state), &maps)

	return maps, err
}

// SetRedisPlayerHuntingPosition - Salvo mappa su redis
func SetRedisMapHunting(maps *pb.Maps) (err error) {
	var jsonValue []byte
	jsonValue, err = json.Marshal(maps)
	if err != nil {
		return err
	}

	err = services.Redis.Set(fmt.Sprintf("hunting_map_%v", maps.ID), string(jsonValue), 0).Err()
	return
}

// CheckState - Verifica ed effettua controlli sullo stato del player in un determinato controller
func CheckState(player pb.Player, controller string, payload interface{}, father uint32) (playerState *pb.PlayerState, isNewState bool, err error) {
	// Filtro gli stati del player recuperando lo stato appartente a questa specifica rotta
	playerState, _ = GetPlayerStateByFunction(player.GetStates(), controller)

	// Non ho trovato nessuna corrispondenza creo una nuova
	if playerState == nil {
		jsonPayload, _ := json.Marshal(payload)

		// Creo il nuovo stato
		rCreatePlayerState, err := services.NnSDK.CreatePlayerState(NewContext(1), &pb.CreatePlayerStateRequest{
			PlayerState: &pb.PlayerState{
				PlayerID:   player.GetID(),
				Controller: controller,
				Father:     father,
				Payload:    string(jsonPayload),
			},
		})

		// Ritoro errore se non riesco a creare lo stato
		if err != nil {
			return playerState, true, err
		}

		// Ritorno stato aggiornato
		playerState = rCreatePlayerState.GetPlayerState()

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

// UnmarshalPayload - Unmarshal payload state
func UnmarshalPayload(payload string, funcInterface interface{}) {
	if payload != "" {
		err := json.Unmarshal([]byte(payload), &funcInterface)
		if err != nil {
			services.ErrorHandler("Error unmarshal payload", err)
		}
	}
}
