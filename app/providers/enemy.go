package providers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type EnemyProvider struct {
	BaseProvider
}

func (ep *EnemyProvider) GetEnemyByID(id uint) (response nnsdk.Enemy, err error) {
	ep.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("enemies/%v", id), nil).Get()
	if err != nil {
		return response, err
	}

	err = ep.Response(&response)
	return
}

func (ep *EnemyProvider) HitEnemy(enemy nnsdk.Enemy, request nnsdk.HitEnemyRequest) (response nnsdk.HitEnemyResponse, err error) {
	ep.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("enemies/%v/hit", enemy.ID), request).Post()
	if err != nil {
		return response, err
	}

	err = ep.Response(&response)
	return
}
