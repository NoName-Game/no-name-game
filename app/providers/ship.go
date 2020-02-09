package providers

import (
	"encoding/json"
	"errors"
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type ResponseExplorationInfo struct {
	Star     nnsdk.Planet
	Distance float64
	Fuel     float64
	Time     float64
}

type RequestExplorationEnd struct {
	Position []float64
	Tank     float64
}

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
		services.ErrorHandler("Can't call ship repairs", err)
		return info, err
	}

	// Verifico se sono ritornati degli errori dalla chiamata
	if resp.Error != "" {
		return info, errors.New(resp.Message)
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		services.ErrorHandler("Can't unmarshal ship repairs", err)
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

func GetShipExplorationInfo(ship nnsdk.Ship) ([]ResponseExplorationInfo, error) {
	// var info []map[string]interface{}
	var info []ResponseExplorationInfo

	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("ships/%v/explorations/info", ship.ID), nil).Get()
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(resp.Data, &info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func EndShipExploration(ship nnsdk.Ship, request RequestExplorationEnd) (map[string]interface{}, error) {
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
