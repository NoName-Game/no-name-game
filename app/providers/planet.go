package providers

import (
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type PlanetProvider struct {
	BaseProvider
}

// GetPlanetByCoordinate - Recupera pianeta da coordinate
func (pp *PlanetProvider) GetPlanetByCoordinate(x float64, y float64, z float64) (response nnsdk.Planet, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("search/planet/coordinate?x=%v&y=%v&z=%v", x, y, z), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}
