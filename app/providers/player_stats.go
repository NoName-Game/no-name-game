package providers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type PlayerStatsProvider struct {
	BaseProvider
}

// TODO: questa funzione non dovrebbe essere attiva, spostare la logica di chi usa questo metodo sul WS
func (pp *PlayerStatsProvider) UpdatePlayerStats(request nnsdk.PlayerStats) (response nnsdk.PlayerStats, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("player-stats/%v", request.ID), request).Patch()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}
