package providers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type MapProvider struct {
	BaseProvider
}

func (mp *MapProvider) GetMapByID(id uint) (response nnsdk.Map, err error) {
	mp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("maps/%v", id), nil).Get()
	if err != nil {
		return response, err
	}

	err = mp.Response(&response)
	return
}
