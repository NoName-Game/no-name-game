package providers

import (
	"encoding/json"
	"fmt"
	"net/url"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func GetItemByID(id uint) (nnsdk.Item, error) {
	var item nnsdk.Item
	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("items/%v", id), nil).Get()
	if err != nil {
		return item, err
	}

	err = json.Unmarshal(resp.Data, &item)
	if err != nil {
		return item, err
	}

	return item, nil
}

func GetAllItems() (nnsdk.Items, error) {
	var items nnsdk.Items
	resp, err := services.NnSDK.MakeRequest("items", nil).Get()
	if err != nil {
		return items, err
	}

	err = json.Unmarshal(resp.Data, &items)
	if err != nil {
		return items, err
	}

	return items, nil
}

func GetItemByName(name string) (nnsdk.Item, error) {
	var item nnsdk.Item
	params := url.Values{}
	params.Add("name", name)

	resp, err := services.NnSDK.MakeRequest("search/item?"+params.Encode(), nil).Get()

	if err != nil {
		return item, err
	}

	err = json.Unmarshal(resp.Data, &item)

	if err != nil {
		return item, err
	}
	return item, nil

}

func GetItemByCategoryID(categoryID uint) (nnsdk.Items, error) {
	var items nnsdk.Items
	params := url.Values{}
	params.Add("category_id", fmt.Sprintf("%v", categoryID))

	resp, err := services.NnSDK.MakeRequest("search/item?"+params.Encode(), nil).Get()

	if err != nil {
		return items, err
	}

	err = json.Unmarshal(resp.Data, &items)

	if err != nil {
		return items, err
	}
	return items, nil

}
