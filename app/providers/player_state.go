package providers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func GetPlayerStateByID(id uint) (nnsdk.PlayerState, error) {
	var playerState nnsdk.PlayerState
	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("player-states/%v", id), nil).Get()
	if err != nil {
		return playerState, err
	}

	err = json.Unmarshal(resp.Data, &playerState)
	if err != nil {
		return playerState, err
	}

	return playerState, nil
}

func GetPlayerStateToNotify() (nnsdk.PlayerStates, error) {
	var playerStates nnsdk.PlayerStates

	resp, err := services.NnSDK.MakeRequest("player-states/to-notify", nil).Get()
	if err != nil {
		return playerStates, err
	}

	err = json.Unmarshal(resp.Data, &playerStates)
	if err != nil {
		return playerStates, err
	}

	return playerStates, nil
}

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

func DeletePlayerState(request nnsdk.PlayerState) (nnsdk.PlayerState, error) {
	var playerState nnsdk.PlayerState
	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("player-states/%v", request.ID), request).Delete()
	if err != nil {
		return playerState, err
	}

	err = json.Unmarshal(resp.Data, &playerState)
	if err != nil {
		return playerState, err
	}

	return playerState, nil
}
