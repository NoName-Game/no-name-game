package providers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type ArmorCategoryProvider struct {
	BaseProvider
}

func (ac *ArmorCategoryProvider) GetAllArmorCategory() (response nnsdk.ArmorCategories, err error) {
	ac.SDKResp, err = services.NnSDK.MakeRequest("armor/categories", nil).Get()
	if err != nil {
		return response, err
	}

	err = ac.Response(&response)
	return
}
