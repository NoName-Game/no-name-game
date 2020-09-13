package helpers

import (
	"encoding/json"
	"errors"
	"strconv"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/services"
)

func CheckStateNew(playerID uint32, controller string, stage int32) (controllerCached string, stageCached int32, err error) {
	// Aggiorno sempre il controller in cui si trova il player
	SetCacheState(playerID, controller)

	// Verifico se sono presenti degli stati per questo controller in memoria
	var playerControllerStateString string
	playerControllerStateString, _ = GetCacheControllerStage(playerID, controller)

	// Se ho trovato qualcosa splitto
	if playerControllerStateString != "" {
		// Recupero stage
		var stage64 int64
		if stage64, err = strconv.ParseInt(playerControllerStateString, 10, 32); err != nil {
			panic(err)
		}

		return controller, int32(stage64), err
	}

	// Se non ho trovato nuessuno stato lo creo
	SetCacheControllerStage(playerID, controller, stage)

	return controller, stage, err
}

// CheckState - Verifica ed effettua controlli sullo stato del player in un determinato controller
func CheckState(player *pb.Player, activeStates []*pb.PlayerState, controller string, payload interface{}, father uint32) (playerState *pb.PlayerState, isNewState bool, err error) {
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

	// Se tutto ok imposto e setto il nuovo stato in cache
	SetCacheState(player.ID, controller)

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
