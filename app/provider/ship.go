package provider

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

func GetShipRepairInfo(ship nnsdk.Ship) (map[string]interface{}, error) {
	var info map[string]interface{}

	resp, err := services.NnSDK.MakeRequest("ships/"+strconv.FormatUint(uint64(ship.ID), 10)+"/repair", nil).Get()
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func PostShipRepair(ship nnsdk.Ship) (map[string]interface{}, error) {
	var info map[string]interface{}

	resp, err := services.NnSDK.MakeRequest("ships/"+strconv.FormatUint(uint64(ship.ID), 10)+"/repair", nil).Post()
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}
