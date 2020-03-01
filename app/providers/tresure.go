package providers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type TresureProvider struct {
	BaseProvider
}

func (tp *TresureProvider) DropTresure(request nnsdk.TresureDropRequest) (response nnsdk.DropResponse, err error) {
	tp.SDKResp, err = services.NnSDK.MakeRequest("tresures/drop", request).Post()
	if err != nil {
		return response, err
	}

	err = tp.Response(&response)
	return
}
