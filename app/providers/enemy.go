package providers

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/services"
)

func GetEnemyByID(id uint) (nnsdk.Enemy, error) {
	var enemy nnsdk.Enemy
	resp, err := services.NnSDK.MakeRequest("enemies/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return enemy, err
	}

	err = json.Unmarshal(resp.Data, &enemy)
	if err != nil {
		return enemy, err
	}

	return enemy, nil
}

func UpdateEnemy(request nnsdk.Enemy) (nnsdk.Enemy, error) {
	var enemy nnsdk.Enemy

	resp, err := services.NnSDK.MakeRequest("enemies/"+strconv.FormatUint(uint64(request.ID), 10), request).Patch()
	if err != nil {
		return enemy, err
	}

	err = json.Unmarshal(resp.Data, &enemy)
	if err != nil {
		return enemy, err
	}

	return enemy, nil
}

func DeleteEnemy(id uint) (nnsdk.Enemy, error) {
	var enemy nnsdk.Enemy
	resp, err := services.NnSDK.MakeRequest("enemies/"+strconv.FormatUint(uint64(id), 10), nil).Delete()
	if err != nil {
		return enemy, err
	}

	err = json.Unmarshal(resp.Data, &enemy)
	if err != nil {
		return enemy, err
	}

	return enemy, nil
}

func Spawn(request nnsdk.Enemy) (nnsdk.Enemy, error) {
	var enemy nnsdk.Enemy
	resp, err := services.NnSDK.MakeRequest("enemies/spawn", request).Post()
	if err != nil {
		return enemy, err
	}

	err = json.Unmarshal(resp.Data, &enemy)
	if err != nil {
		return enemy, err
	}

	return enemy, nil
}

func EnemyDamage(id uint) (float64, error) {
	var damage float64
	resp, err := services.NnSDK.MakeRequest("enemies/"+strconv.FormatUint(uint64(id), 10)+"/damage", nil).Get()
	if err != nil {
		return 0, err
	}

	err = json.Unmarshal(resp.Data, &damage)
	if err != nil {
		return 0, err
	}
	return damage, nil
}
