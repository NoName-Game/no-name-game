package provider

import (
	"encoding/json"
	"fmt"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

func GetShipRepairInfo(ship nnsdk.Ship) (map[string]interface{}, error) {
	var info map[string]interface{}

	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/repairs/info", ship.ID), nil).Get()
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func StartShipRepair(ship nnsdk.Ship) (map[uint]interface{}, error) {
	var info map[uint]interface{}

	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/repairs/start", ship.ID), nil).Post()
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func EndShipRepair(ship nnsdk.Ship) (map[string]interface{}, error) {
	var info map[string]interface{}

	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/repairs/end", ship.ID), nil).Post()
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}
