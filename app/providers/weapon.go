package providers

import (
	"fmt"
	"net/url"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type WeaponProvider struct {
	BaseProvider
}

func (wp *WeaponProvider) GetWeaponByID(id uint) (response nnsdk.Weapon, err error) {
	wp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("weapons/%v", id), nil).Get()
	if err != nil {
		return response, err
	}

	err = wp.Response(&response)
	return
}

func (wp *WeaponProvider) FindWeaponByName(name string) (response nnsdk.Weapon, err error) {
	params := url.Values{}
	params.Add("name", name)

	wp.SDKResp, err = services.NnSDK.MakeRequest("search/weapon?"+params.Encode(), nil).Get()
	if err != nil {
		return response, err
	}

	err = wp.Response(&response)
	return
}

func (wp *WeaponProvider) UpdateWeapon(request nnsdk.Weapon) (response nnsdk.Weapon, err error) {
	wp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("weapons/%v", request.ID), request).Patch()
	if err != nil {
		return response, err
	}

	err = wp.Response(&response)
	return
}
