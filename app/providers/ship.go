package providers

import (
	"encoding/json"
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type ShipProvider struct {
	BaseProvider
}

func (sp *ShipProvider) GetShipRepairInfo(ship nnsdk.Ship) (response nnsdk.ShipRepairInfoResponse, err error) {
	sp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/repairs/info", ship.ID), nil).Get()
	if err != nil {
		return response, err
	}

	err = sp.Response(&response)
	return
}

func (sp *ShipProvider) StartShipRepair(ship nnsdk.Ship) (response []nnsdk.ShipRepairStartResponse, err error) {
	sp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/repairs/start", ship.ID), nil).Post()
	if err != nil {
		return response, err
	}

	err = sp.Response(&response)
	return
}

func (sp *ShipProvider) EndShipRepair(ship nnsdk.Ship) (err error) {
	_, err = services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/repairs/end", ship.ID), nil).Post()

	return
}

func (sp *ShipProvider) GetShipExplorationInfo(ship nnsdk.Ship) (response []nnsdk.ExplorationInfoResponse, err error) {
	sp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/explorations/info", ship.ID), nil).Get()
	if err != nil {
		return response, err
	}

	err = sp.Response(&response)
	return
}

func (sp *ShipProvider) EndShipExploration(ship nnsdk.Ship, request nnsdk.ExplorationEndRequest) (map[string]interface{}, error) {
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
