package provider

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

func GetWeaponByID(id uint) (nnsdk.Weapon, error) {
	var weapon nnsdk.Weapon
	resp, err := services.NnSDK.MakeRequest("weapons/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return weapon, err
	}

	err = json.Unmarshal(resp.Data, &weapon)
	if err != nil {
		return weapon, err
	}

	return weapon, nil
}

func FindWeaponByName(name string) (nnsdk.Weapon, error) {
	var weapon nnsdk.Weapon
	resp, err := services.NnSDK.MakeRequest("search/weapon?name="+name, nil).Get()
	if err != nil {
		return weapon, err
	}

	err = json.Unmarshal(resp.Data, &weapon)
	if err != nil {
		return weapon, err
	}

	return weapon, nil
}

func UpdateWeapon(request nnsdk.Weapon) (nnsdk.Weapon, error) {
	var weapon nnsdk.Weapon
	resp, err := services.NnSDK.MakeRequest("weapons/"+strconv.FormatUint(uint64(request.ID), 10), request).Patch()
	if err != nil {
		return weapon, err
	}

	err = json.Unmarshal(resp.Data, &weapon)
	if err != nil {
		return weapon, err
	}

	return weapon, nil
}

func DeleteWeapon(request nnsdk.Weapon) (nnsdk.Weapon, error) {
	var weapon nnsdk.Weapon
	resp, err := services.NnSDK.MakeRequest("weapons/"+strconv.FormatUint(uint64(request.ID), 10), request).Delete()
	if err != nil {
		return weapon, err
	}

	err = json.Unmarshal(resp.Data, &weapon)
	if err != nil {
		return weapon, err
	}

	return weapon, nil
}
