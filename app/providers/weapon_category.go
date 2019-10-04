package providers

import (
	"encoding/json"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func GetAllWeaponCategory() (nnsdk.WeaponCategories, error) {
	var categories nnsdk.WeaponCategories
	resp, err := services.NnSDK.MakeRequest("weapon/categories", nil).Get()
	if err != nil {
		return categories, err
	}

	err = json.Unmarshal(resp.Data, &categories)
	if err != nil {
		return categories, err
	}

	return categories, nil
}
