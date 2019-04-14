package provider

import (
	"encoding/json"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

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
