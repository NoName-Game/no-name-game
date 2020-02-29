package providers

import (
	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type ItemCategoryProvider struct {
	BaseProvider
}

func (ip *ItemCategoryProvider) GetAllItemCategories() (response nnsdk.ItemCategories, err error) {
	ip.SDKResp, err = services.NnSDK.MakeRequest("item/categories", nil).Get()
	if err != nil {
		return response, err
	}

	err = ip.Response(&response)
	return
}
