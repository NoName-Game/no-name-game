package providers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
	"encoding/json"
	"strconv"
)

func GetRecipeByID(id uint) (nnsdk.Recipe, error) {
	var resource nnsdk.Recipe
	resp, err := services.NnSDK.MakeRequest("recipe/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return resource, err
	}

	err = json.Unmarshal(resp.Data, &resource)
	if err != nil {
		return resource, err
	}

	return resource, nil
}

func GetAllRecipe() (nnsdk.Recipes, error) {
	var resource nnsdk.Recipes
	resp, err := services.NnSDK.MakeRequest("recipe/", nil).Get()
	if err != nil {
		return resource, err
	}

	err = json.Unmarshal(resp.Data, &resource)
	if err != nil {
		return resource, err
	}

	return resource, nil
}
