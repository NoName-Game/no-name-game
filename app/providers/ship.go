package providers

import (
	"encoding/json"
	"errors"
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func GetShipRepairInfo(ship nnsdk.Ship) (info nnsdk.ShipRepairInfoResponse, err error) {
	var resp nnsdk.APIResponse
	resp, err = services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/repairs/info", ship.ID), nil).Get()
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func StartShipRepair(ship nnsdk.Ship) (info []nnsdk.ShipRepairStartResponse, err error) {
	var resp nnsdk.APIResponse
	resp, err = services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/repairs/start", ship.ID), nil).Post()
	if err != nil {
		return info, err
	}

	if resp.Error != "" {
		err = errors.New(resp.Error)

		return info, err
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func EndShipRepair(ship nnsdk.Ship) (err error) {
	_, err = services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/repairs/end", ship.ID), nil).Post()

	return
}

func GetShipExplorationInfo(ship nnsdk.Ship) (info []nnsdk.ResponseExplorationInfo, err error) {
	var resp nnsdk.APIResponse
	resp, err = services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/explorations/info", ship.ID), nil).Get()
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func EndShipExploration(ship nnsdk.Ship, request nnsdk.RequestExplorationEnd) (map[string]interface{}, error) {
	var info map[string]interface{}

	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/explorations/end", ship.ID), request).Post()
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}
