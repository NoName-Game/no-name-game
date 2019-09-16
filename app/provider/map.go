package provider

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

func GetMapByID(id uint) (nnsdk.Map, error) {
	var Map nnsdk.Map
	resp, err := services.NnSDK.MakeRequest("maps/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return Map, err
	}

	err = json.Unmarshal(resp.Data, &Map)
	if err != nil {
		return Map, err
	}

	return Map, nil
}

func CreateMap(playerID uint) (nnsdk.Map, error) {
	var Map nnsdk.Map
	Map.PlayerID = playerID
	resp, err := services.NnSDK.MakeRequest("maps", Map).Post()
	if err != nil {
		return Map, err
	}

	err = json.Unmarshal(resp.Data, &Map)
	if err != nil {
		return Map, err
	}

	return Map, nil
}

func UpdateMap(request nnsdk.Map) (nnsdk.Map, error) {
	var Map nnsdk.Map

	resp, err := services.NnSDK.MakeRequest("maps/"+strconv.FormatUint(uint64(request.ID), 10), request).Patch()
	if err != nil {
		return Map, err
	}

	err = json.Unmarshal(resp.Data, &Map)
	if err != nil {
		return Map, err
	}

	return Map, nil
}
