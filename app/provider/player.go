package provider

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

func GetPlayerByID(id uint) (nnsdk.Player, error) {
	var player nnsdk.Player
	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return player, err
	}

	err = json.Unmarshal(resp.Data, &player)
	if err != nil {
		return player, err
	}

	return player, nil
}

func FindPlayerByUsername(username string) (nnsdk.Player, error) {
	var player nnsdk.Player
	resp, err := services.NnSDK.MakeRequest("search/player?username="+username, nil).Get()
	if err != nil {
		return player, err
	}

	err = json.Unmarshal(resp.Data, &player)
	if err != nil {
		return player, err
	}

	return player, nil
}

func CreatePlayer(request nnsdk.Player) (nnsdk.Player, error) {
	var player nnsdk.Player
	resp, err := services.NnSDK.MakeRequest("players", request).Post()
	if err != nil {
		return player, err
	}

	err = json.Unmarshal(resp.Data, &player)
	if err != nil {
		return player, err
	}

	return player, nil
}

func GetPlayerStates(player nnsdk.Player) (nnsdk.PlayerStates, error) {
	var playerStates nnsdk.PlayerStates

	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(player.ID), 10)+"/states", nil).Get()
	if err != nil {
		return playerStates, err
	}

	err = json.Unmarshal(resp.Data, &playerStates)
	if err != nil {
		return playerStates, err
	}

	return playerStates, nil
}
