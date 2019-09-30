package providers

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func GetResourceByID(id uint) (nnsdk.Resource, error) {
	var resource nnsdk.Resource
	resp, err := services.NnSDK.MakeRequest("resources/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return resource, err
	}

	err = json.Unmarshal(resp.Data, &resource)
	if err != nil {
		return resource, err
	}

	return resource, nil
}

func GetRandomResource(categoryID uint) (nnsdk.Resource, error) {
	var resource nnsdk.Resource
	resp, err := services.NnSDK.MakeRequest("resources/drop/"+strconv.FormatUint(uint64(categoryID), 10), nil).Get()
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
	resp, err := services.NnSDK.MakeRequest("search/resource?name="+name, nil).Get()
	if err != nil {
		return resource, err
	}

	err = json.Unmarshal(resp.Data, &resource)
	if err != nil {
		return resource, err
	}

	return resource, nil
}
