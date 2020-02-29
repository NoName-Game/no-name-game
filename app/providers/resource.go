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

func (rp *ResourceProvider) DropResource(typeExploration string, qtyExploration int, playerID uint, planetID uint) (response nnsdk.DropItem, err error) {
	// TODO spostare in reuqest
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

	rp.SDKResp, err = services.NnSDK.MakeRequest("resources/drop", request).Post()
	if err != nil {
		return response, err
	}

	err = rp.Response(&response)
	return
}
