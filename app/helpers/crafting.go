package helpers

import (
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"
)

// ListRecipe - Metodo che aiuta a recuperare la lista di risore necessarie
// al crafting di un determianto item
func ListRecipe(needed map[uint]int) (result string, err error) {
	var resourceProvider providers.ResourceProvider
	for resourceID, value := range needed {
		var resource nnsdk.Resource
		resource, err = resourceProvider.GetResourceByID(resourceID)
		if err != nil {
			return result, err
		}

		result += resource.Name + " x" + strconv.Itoa(value) + "\n"
	}

	return result, err
}
