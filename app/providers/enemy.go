package providers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
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

func HitEnemy(enemy nnsdk.Enemy, request nnsdk.HitEnemyRequest) (response nnsdk.HitEnemyResponse, err error) {
	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("enemies/%v/hit", enemy.ID), request).Post()
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(resp.Data, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}
