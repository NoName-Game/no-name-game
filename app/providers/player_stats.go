package providers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func CreatePlayerStats(request nnsdk.PlayerStats) (nnsdk.PlayerStats, error) {
	var playerStats nnsdk.PlayerStats
	resp, err := services.NnSDK.MakeRequest("player-stats", request).Post()
	if err != nil {
		return playerStats, err
	}

	err = json.Unmarshal(resp.Data, &playerStats)
	if err != nil {
		return playerStats, err
	}

	return playerStats, nil
}

// TODO: questa funzione non dovrebbe essere attiva, spostare la logica di chi usa questo metodo sul WS
func UpdatePlayerStats(request nnsdk.PlayerStats) (nnsdk.PlayerStats, error) {
	var playerStats nnsdk.PlayerStats
	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("player-stats/%v", request.ID), request).Patch()
	if err != nil {
		return playerStats, err
	}

	err = json.Unmarshal(resp.Data, &playerStats)
	if err != nil {
		return playerStats, err
	}

	return playerStats, nil
}

func DeletePlayerStats(request nnsdk.PlayerStats) (nnsdk.PlayerStats, error) {
	var playerStats nnsdk.PlayerStats
	resp, err := services.NnSDK.MakeRequest("player-stats/"+strconv.FormatUint(uint64(request.ID), 10), request).Delete()
	if err != nil {
		return playerStats, err
	}

	err = json.Unmarshal(resp.Data, &playerStats)
	if err != nil {
		return playerStats, err
	}

	return playerStats, nil
}
