package providers

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
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

func DeleteMap(request nnsdk.Map) (nnsdk.Map, error) {
	var Map nnsdk.Map

	resp, err := services.NnSDK.MakeRequest("maps/"+strconv.FormatUint(uint64(request.ID), 10), request).Delete()
	if err != nil {
		return Map, err
	}

	err = json.Unmarshal(resp.Data, &Map)
	if err != nil {
		return Map, err
	}

	return Map, nil
}
