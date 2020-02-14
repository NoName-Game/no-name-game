package providers

import (
	"encoding/json"
	"fmt"
	"net/url"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func GetResourceByID(id uint) (nnsdk.Resource, error) {
	var resource nnsdk.Resource
	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("resources/%v", id), nil).Get()
	if err != nil {
		return resource, err
	}

	err = json.Unmarshal(resp.Data, &resource)
	if err != nil {
		return resource, err
	}

	return resource, nil
}

func DropResource(typeExploration string, qtyExploration int, playerID uint, planetID uint) (nnsdk.DropItem, error) {
	type dropRequest struct {
		TypeExploration string
		QtyExploration  int
		PlayerID        uint
		PlanetID        uint
	}

	request := dropRequest{
		TypeExploration: typeExploration,
		QtyExploration:  qtyExploration,
		PlayerID:        playerID,
		PlanetID:        planetID,
	}

	var resource nnsdk.DropItem
	resp, err := services.NnSDK.MakeRequest("resources/drop", request).Post()
	if err != nil {
		return resource, err
	}

	err = json.Unmarshal(resp.Data, &resource)
	if err != nil {
		return resource, err
	}

	return resource, nil
}

func FindResourceByName(name string) (nnsdk.Resource, error) {
	var resource nnsdk.Resource

	// Encode paramiters
	params := url.Values{}
	params.Add("name", name)

	resp, err := services.NnSDK.MakeRequest("search/resource?"+params.Encode(), nil).Get()
	if err != nil {
		return resource, err
	}

	err = json.Unmarshal(resp.Data, &resource)
	if err != nil {
		return resource, err
	}

	return resource, nil
}
