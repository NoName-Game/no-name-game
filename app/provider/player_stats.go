package provider

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
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

func UpdatePlayerStats(request nnsdk.PlayerStats) (nnsdk.PlayerStats, error) {
	var playerStats nnsdk.PlayerStats
	resp, err := services.NnSDK.MakeRequest("player-stats/"+strconv.FormatUint(uint64(request.ID), 10), request).Patch()
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
