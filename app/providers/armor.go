package providers

import (
	"fmt"
	"net/url"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type ArmorProvider struct {
	BaseProvider
}

func (ap *ArmorProvider) GetArmorByID(id uint) (response nnsdk.Armor, err error) {
	ap.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("armors/%v", id), nil).Get()
	if err != nil {
		return response, err
	}

	err = ap.Response(&response)
	return
}

func (ap *ArmorProvider) FindArmorByName(name string) (response nnsdk.Armor, err error) {
	// Encode paramiters
	params := url.Values{}
	params.Add("name", name)

	ap.SDKResp, err = services.NnSDK.MakeRequest("search/armor?"+params.Encode(), nil).Get()
	if err != nil {
		return response, err
	}

	err = ap.Response(&response)
	return
}

func (ap *ArmorProvider) UpdateArmor(request nnsdk.Armor) (response nnsdk.Armor, err error) {
	ap.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("armors/%v", request.ID), request).Patch()
	if err != nil {
		return response, err
	}

	err = ap.Response(&response)
	return
}

// TODO: Da verificare
// func CraftArmor(request nnsdk.ArmorCraft) (nnsdk.Armor, error) {
// 	var armor nnsdk.Armor
// 	resp, err := services.NnSDK.MakeRequest("armors/craft", request).Post()
// 	if err != nil {
// 		return armor, err
// 	}
//
// 	err = json.Unmarshal(resp.Data, &armor)
// 	if err != nil {
// 		return armor, err
// 	}
//
// 	return armor, nil
// }
