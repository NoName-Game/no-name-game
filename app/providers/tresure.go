package providers

import (
	"encoding/json"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func DropTresure(playerID uint, tresureID uint) (nnsdk.DropResponse, error) {
	request := nnsdk.TresureDropRequest{
		PlayerID:  playerID,
		TresureID: tresureID,
	}

	var drop nnsdk.DropResponse
	resp, err := services.NnSDK.MakeRequest("tresures/drop", request).Post()
	if err != nil {
		return drop, err
	}

	err = json.Unmarshal(resp.Data, &drop)
	if err != nil {
		return drop, err
	}

	return drop, nil
}
