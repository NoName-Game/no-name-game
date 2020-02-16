package providers

import (
	"encoding/json"
	"fmt"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

// GetPlanetByCoordinate - Recupera pianeta da coordinate
func GetPlanetByCoordinate(x float64, y float64, z float64) (nnsdk.Planet, error) {
	var planet nnsdk.Planet

	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("search/planet/coordinate?x=%v&y=%v&z=%v", x, y, z), nil).Get()
	if err != nil {
		return planet, err
	}

	err = json.Unmarshal(resp.Data, &planet)
	if err != nil {
		return planet, err
	}

	return planet, nil
}
