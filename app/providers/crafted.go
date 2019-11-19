package providers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
	"encoding/json"
	"strconv"
)

func GetCraftedByID(id uint) (nnsdk.Crafted, error) {
	var resource nnsdk.Crafted
	resp, err := services.NnSDK.MakeRequest("craftable/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return resource, err
	}

	err = json.Unmarshal(resp.Data, &resource)
	if err != nil {
		return resource, err
	}

	return resource, nil
}
