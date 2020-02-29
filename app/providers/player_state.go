package providers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type PlayerStateProvider struct {
	BaseProvider
}

func (pp *PlayerStateProvider) GetPlayerStateByID(id uint) (response nnsdk.PlayerState, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("player-states/%v", id), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerStateProvider) GetPlayerStateToNotify() (response nnsdk.PlayerStates, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest("player-states/to-notify", nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerStateProvider) CreatePlayerState(request nnsdk.PlayerState) (response nnsdk.PlayerState, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest("player-states", request).Post()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerStateProvider) UpdatePlayerState(request nnsdk.PlayerState) (response nnsdk.PlayerState, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("player-states/%v", request.ID), request).Patch()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerStateProvider) DeletePlayerState(request nnsdk.PlayerState) (response nnsdk.PlayerState, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("player-states/%v", request.ID), request).Delete()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}
