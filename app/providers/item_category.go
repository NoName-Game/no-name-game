package providers

import (
	"encoding/json"
	"errors"
	"net/url"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func GetAllItemCategories() (itemCategories nnsdk.ItemCategories, err error) {
	var resp nnsdk.APIResponse
	resp, err = services.NnSDK.MakeRequest("item/categories", nil).Get()
	if err != nil {
		return itemCategories, err
	}

	err = json.Unmarshal(resp.Data, &itemCategories)
	if err != nil {
		return itemCategories, err
	}

	return itemCategories, nil
}

func FindItemCategoryByName(name string) (itemCategory nnsdk.ItemCategory, err error) {
	// Encode paramiters
	params := url.Values{}
	params.Add("name", name)

	var resp nnsdk.APIResponse
	resp, err = services.NnSDK.MakeRequest("search/item/category?"+params.Encode(), nil).Get()
	if err != nil {
		return itemCategory, err
	}

	// TODO: da gestire meglio
	// Verifico il tipo di risposta
	if resp.StatusCode == 404 {
		return itemCategory, err
	} else if resp.StatusCode == 400 {
		err = errors.New(resp.Error)
		return itemCategory, err
	}

	err = json.Unmarshal(resp.Data, &itemCategory)
	if err != nil {
		return itemCategory, err
	}

	return itemCategory, nil
}
