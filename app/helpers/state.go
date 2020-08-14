package helpers

import (
	"encoding/json"
	"errors"
	"fmt"

	gocache "github.com/patrickmn/go-cache"

	pb "bitbucket.org/no-name-game/nn-grpc/rpc"

	"bitbucket.org/no-name-game/nn-telegram/services"
)

// GetRedisState - Metodo generico per il recupero degli stati di un player
func GetCacheState(playerID uint32) (controller string, err error) {
	value, found := services.Cache.Get(fmt.Sprintf("state_%v", playerID))
	if found {
		return value.(string), nil
	}

	err = errors.New("cached state not found")
	return
}

// SetCacheState - Metodo generico per il settaggio di uno stato in memoria di un determinato player
func SetCacheState(playerID uint32, data string) {
	services.Cache.Set(
		fmt.Sprintf("state_%v", playerID),
		data,
		gocache.NoExpiration,
	)
}

// DelRedisState - Metodo generico per la cancellazione degli stati di un determinato player
func DelCacheState(playerID uint32) {
	services.Cache.Delete(fmt.Sprintf("state_%v", playerID))
	return
}

// CheckState - Verifica ed effettua controlli sullo stato del player in un determinato controller
func CheckState(player pb.Player, activeStates []*pb.PlayerState, controller string, payload interface{}, father uint32) (playerState *pb.PlayerState, isNewState bool, err error) {
	// Filtro gli stati del player recuperando lo stato appartente a questa specifica rotta
	playerState, _ = GetPlayerStateByFunction(activeStates, controller)

	// Non ho trovato nessuna corrispondenza creo una nuova
	if playerState == nil {
		jsonPayload, _ := json.Marshal(payload)

		// Creo il nuovo stato
		var rCreatePlayerState *pb.CreatePlayerStateResponse
		rCreatePlayerState, err = services.NnSDK.CreatePlayerState(NewContext(1), &pb.CreatePlayerStateRequest{
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
	SetCacheState(player.ID, controller)

	return
}
