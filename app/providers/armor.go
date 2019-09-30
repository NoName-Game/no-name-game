package providers

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

func GetArmorByID(id uint) (nnsdk.Armor, error) {
	var armor nnsdk.Armor
	resp, err := services.NnSDK.MakeRequest("armors/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return armor, err
	}

	err = json.Unmarshal(resp.Data, &armor)
	if err != nil {
		return armor, err
	}

	return armor, nil
}

func FindArmorByName(name string) (nnsdk.Armor, error) {
	var armor nnsdk.Armor
	resp, err := services.NnSDK.MakeRequest("search/armor?name="+name, nil).Get()
	if err != nil {
		return armor, err
	}

	err = json.Unmarshal(resp.Data, &armor)
	if err != nil {
		return armor, err
	}

	return armor, nil
}

func UpdateArmor(request nnsdk.Armor) (nnsdk.Armor, error) {
	var armor nnsdk.Armor
	resp, err := services.NnSDK.MakeRequest("armors/"+strconv.FormatUint(uint64(request.ID), 10), request).Patch()
	if err != nil {
		return armor, err
	}

	err = json.Unmarshal(resp.Data, &armor)
	if err != nil {
		return armor, err
	}

	return armor, nil
}

func DeleteArmor(request nnsdk.Armor) (nnsdk.Armor, error) {
	var armor nnsdk.Armor
	resp, err := services.NnSDK.MakeRequest("armors/"+strconv.FormatUint(uint64(request.ID), 10), request).Delete()
	if err != nil {
		return armor, err
	}

	err = json.Unmarshal(resp.Data, &armor)
	if err != nil {
		return armor, err
	}

	return armor, nil
}

func CraftArmor(request nnsdk.ArmorCraft) (nnsdk.Armor, error) {
	var armor nnsdk.Armor
	resp, err := services.NnSDK.MakeRequest("armors/craft", request).Post()
	if err != nil {
		return armor, err
	}

	err = json.Unmarshal(resp.Data, &armor)
	if err != nil {
		return armor, err
	}

	return armor, nil
}
