package providers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type ResourceProvider struct {
	BaseProvider
}

func (rp *ResourceProvider) GetResourceByID(id uint) (response nnsdk.Resource, err error) {
	rp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("resources/%v", id), nil).Get()
	if err != nil {
		return response, err
	}

	err = rp.Response(&response)
	return
}

func (rp *ResourceProvider) DropResource(request nnsdk.ResourceDropRequest) (response nnsdk.DropItem, err error) {
	rp.SDKResp, err = services.NnSDK.MakeRequest("resources/drop", request).Post()
	if err != nil {
		return response, err
	}

	err = rp.Response(&response)
	return
}
