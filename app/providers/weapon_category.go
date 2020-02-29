package providers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type WeaponCateogoryProvider struct {
	BaseProvider
}

func (wp *WeaponCateogoryProvider) GetAllWeaponCategory() (response nnsdk.WeaponCategories, err error) {
	wp.SDKResp, err = services.NnSDK.MakeRequest("weapon/categories", nil).Get()
	if err != nil {
		return response, err
	}

	err = wp.Response(&response)
	return
}
