package helpers

import (
	"strconv"
)

func CheckState(playerID uint32, controller string, stage int32) (controllerCached string, stageCached int32, err error) {
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
