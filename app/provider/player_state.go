package provider

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

func CreatePlayerState(request nnsdk.PlayerState) (nnsdk.PlayerState, error) {
	var playerState nnsdk.PlayerState
	resp, err := services.NnSDK.MakeRequest("player-states", request).Post()
	if err != nil {
		return playerState, err
	}

	err = json.Unmarshal(resp.Data, &playerState)
	if err != nil {
		return playerState, err
	}

	return playerState, nil
}

func UpdatePlayerState(request nnsdk.PlayerState) (nnsdk.PlayerState, error) {
	var playerState nnsdk.PlayerState
	resp, err := services.NnSDK.MakeRequest("player-states/"+strconv.FormatUint(uint64(request.ID), 10), request).Patch()
	if err != nil {
		return playerState, err
	}

	err = json.Unmarshal(resp.Data, &playerState)
	if err != nil {
		return playerState, err
	}

	return playerState, nil
}
