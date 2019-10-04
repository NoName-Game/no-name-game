package providers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

func GetPlayerByID(id uint) (nnsdk.Player, error) {
	var player nnsdk.Player
	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return player, err
	}

	err = json.Unmarshal(resp.Data, &player)
	if err != nil {
		return player, err
	}

	return player, nil
}

func FindPlayerByUsername(username string) (nnsdk.Player, error) {
	var player nnsdk.Player
	resp, err := services.NnSDK.MakeRequest("search/player?username="+username, nil).Get()
	if err != nil {
		return player, err
	}

	err = json.Unmarshal(resp.Data, &player)
	if err != nil {
		return player, err
	}

	return player, nil
}

func CreatePlayer(request nnsdk.Player) (nnsdk.Player, error) {
	var player nnsdk.Player
	resp, err := services.NnSDK.MakeRequest("players", request).Post()
	if err != nil {
		return player, err
	}

	err = json.Unmarshal(resp.Data, &player)
	if err != nil {
		return player, err
	}

	return player, nil
}

func UpdatePlayer(request nnsdk.Player) (nnsdk.Player, error) {
	var player nnsdk.Player

	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(request.ID), 10), request).Patch()
	if err != nil {
		return player, err
	}

	err = json.Unmarshal(resp.Data, &player)
	if err != nil {
		return player, err
	}

	return player, nil
}

func GetPlayerStates(player nnsdk.Player) (nnsdk.PlayerStates, error) {
	var playerStates nnsdk.PlayerStates

	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(player.ID), 10)+"/states", nil).Get()
	if err != nil {
		return playerStates, err
	}

	err = json.Unmarshal(resp.Data, &playerStates)
	if err != nil {
		return playerStates, err
	}

	return playerStates, nil
}

func GetPlayerStats(player nnsdk.Player) (nnsdk.PlayerStats, error) {
	var playerStats nnsdk.PlayerStats

	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(player.ID), 10)+"/stats", nil).Get()
	if err != nil {
		return playerStats, err
	}

	err = json.Unmarshal(resp.Data, &playerStats)
	if err != nil {
		return playerStats, err
	}

	return playerStats, nil
}

func GetPlayerArmors(player nnsdk.Player, equipped string) (nnsdk.Armors, error) {
	var armors nnsdk.Armors

	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(player.ID), 10)+"/armors?equipped="+equipped, nil).Get()
	if err != nil {
		return armors, err
	}

	err = json.Unmarshal(resp.Data, &armors)
	if err != nil {
		return armors, err
	}

	return armors, nil
}

func GetPlayerWeapons(player nnsdk.Player, equipped string) (nnsdk.Weapons, error) {
	var weapons nnsdk.Weapons

	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(player.ID), 10)+"/weapons?equipped="+equipped, nil).Get()
	if err != nil {
		return weapons, err
	}

	err = json.Unmarshal(resp.Data, &weapons)
	if err != nil {
		return weapons, err
	}

	return weapons, nil
}

func GetPlayerShips(player nnsdk.Player, equipped bool) (nnsdk.Ships, error) {
	var ships nnsdk.Ships

	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/ships?equipped=%v", player.ID, equipped), nil).Get()
	if err != nil {
		return ships, err
	}

	err = json.Unmarshal(resp.Data, &ships)
	if err != nil {
		return ships, err
	}

	return ships, nil
}

func GetPlayerLastPosition(player nnsdk.Player) (nnsdk.PlayerPosition, error) {
	var position nnsdk.PlayerPosition

	resp, err := services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/positions/last", player.ID), nil).Get()
	if err != nil {
		return position, err
	}

	err = json.Unmarshal(resp.Data, &position)
	if err != nil {
		return position, err
	}

	return position, nil
}

func GetPlayerInventory(player nnsdk.Player) (nnsdk.Inventory, error) {
	var inventory nnsdk.Inventory

	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(player.ID), 10)+"/inventory", nil).Get()
	if err != nil {
		return inventory, err
	}

	err = json.Unmarshal(resp.Data, &inventory)
	if err != nil {
		return inventory, err
	}

	return inventory, nil
}

func AddResourceToPlayerInventory(player nnsdk.Player, request nnsdk.AddResourceRequest) (nnsdk.Inventory, error) {
	var inventory nnsdk.Inventory

	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(player.ID), 10)+"/inventory/resource/add", request).Post()
	if err != nil {
		return inventory, err
	}

	err = json.Unmarshal(resp.Data, &inventory)
	if err != nil {
		return inventory, err
	}

	return inventory, nil
}

func RemoveResourceToPlayerInventory(player nnsdk.Player, request nnsdk.AddResourceRequest) (nnsdk.Inventory, error) {
	var inventory nnsdk.Inventory

	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(player.ID), 10)+"/inventory/resource/remove", request).Post()
	if err != nil {
		return inventory, err
	}

	err = json.Unmarshal(resp.Data, &inventory)
	if err != nil {
		return inventory, err
	}

	return inventory, nil
}

func PlayerDamage(id uint) (float64, error) {
	var damage float64
	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(id), 10)+"/damage", nil).Get()
	if err != nil {
		return 0, err
	}

	err = json.Unmarshal(resp.Data, &damage)
	if err != nil {
		return 0, err
	}
	return damage, nil
}

func PlayerPrecision(id uint, selection uint) (float64, error) {
	var damage float64
	resp, err := services.NnSDK.MakeRequest("players/"+strconv.FormatUint(uint64(id), 10)+"/precision/"+strconv.FormatUint(uint64(id), 10), nil).Get()
	if err != nil {
		return 0, err
	}

	err = json.Unmarshal(resp.Data, &damage)
	if err != nil {
		return 0, err
	}
	return damage, nil
}
