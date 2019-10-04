package providers

import (
	"encoding/json"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func GetAllArmorCategory() (nnsdk.ArmorCategories, error) {
	var categories nnsdk.ArmorCategories
	resp, err := services.NnSDK.MakeRequest("armor/categories", nil).Get()
	if err != nil {
		return categories, err
	}

	err = json.Unmarshal(resp.Data, &categories)
	if err != nil {
		return categories, err
	}

	return categories, nil
}
