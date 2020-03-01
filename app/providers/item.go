package providers

import (
	"fmt"
	"net/url"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type ItemProvider struct {
	BaseProvider
}

func (ip *ItemProvider) GetAllItems() (response nnsdk.Items, err error) {
	ip.SDKResp, err = services.NnSDK.MakeRequest("items", nil).Get()
	if err != nil {
		return response, err
	}

	err = ip.Response(&response)
	return
}

func (ip *ItemProvider) GetItemByCategoryID(categoryID uint) (response nnsdk.Items, err error) {
	params := url.Values{}
	params.Add("category_id", fmt.Sprintf("%v", categoryID))

	ip.SDKResp, err = services.NnSDK.MakeRequest("search/item?"+params.Encode(), nil).Get()
	if err != nil {
		return response, err
	}

	err = ip.Response(&response)
	return
}

func (ip *ItemProvider) UseItem(request nnsdk.UseItemRequest) (err error) {
	_, err = services.NnSDK.MakeRequest("items/use", request).Post()
	if err != nil {
		return err
	}

	return nil
}
