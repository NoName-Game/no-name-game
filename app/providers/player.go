package providers

import (
	"fmt"
	"net/url"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/services"
)

type PlayerProvider struct {
	BaseProvider
}

func (pp *PlayerProvider) GetPlayerByID(id uint) (response nnsdk.Player, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v", id), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) FindPlayerByUsername(username string) (response nnsdk.Player, err error) {
	params := url.Values{}
	params.Add("username", username)

	pp.SDKResp, err = services.NnSDK.MakeRequest("search/player?"+params.Encode(), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) CreatePlayer(request nnsdk.Player) (response nnsdk.Player, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest("players", request).Post()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) UpdatePlayer(request nnsdk.Player) (response nnsdk.Player, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v", request.ID), request).Patch()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) GetPlayerStates(player nnsdk.Player) (response nnsdk.PlayerStates, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/states", player.ID), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) GetPlayerStats(player nnsdk.Player) (response nnsdk.PlayerStats, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/stats", player.ID), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) GetPlayerArmors(player nnsdk.Player, equipped string) (response nnsdk.Armors, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/armors?equipped=%v", player.ID, equipped), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) GetPlayerWeapons(player nnsdk.Player, equipped string) (response nnsdk.Weapons, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/weapons?equipped=%v", player.ID, equipped), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) GetPlayerShips(player nnsdk.Player, equipped bool) (response nnsdk.Ships, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/ships?equipped=%v", player.ID, equipped), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

// GetPlayerLastPosition - Recupera ultima posizione del player
func (pp *PlayerProvider) GetPlayerLastPosition(player nnsdk.Player) (response nnsdk.PlayerPosition, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/positions/last", player.ID), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) GetPlayerInventory(player nnsdk.Player) (response nnsdk.PlayerInventories, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/inventory", player.ID), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) SignIn(request nnsdk.Player) (response nnsdk.Player, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest("players/signin", request).Post()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) ManagePlayerInventory(playerID uint, request nnsdk.ManageInventoryRequest) (err error) {
	_, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/inventory/manage", playerID), request).Post()
	if err != nil {
		return err
	}

	return nil
}

func (pp *PlayerProvider) GetPlayerResources(playerID uint) (response nnsdk.PlayerInventories, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/inventory/resources", playerID), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) GetPlayerItems(playerID uint) (response nnsdk.PlayerInventories, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/inventory/items", playerID), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) GetPlayerEconomy(playerID uint, economyType string) (response nnsdk.MoneyResponse, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/%s", playerID, economyType), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) GetRestsInfo(playerID uint) (response nnsdk.PlayerRestInfoResponse, err error) {
	pp.SDKResp, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/rests/info", playerID), nil).Get()
	if err != nil {
		return response, err
	}

	err = pp.Response(&response)
	return
}

func (pp *PlayerProvider) EndPlayerRest(playerID uint, request nnsdk.PlayerRestEndRequest) (err error) {
	_, err = services.NnSDK.MakeRequest(fmt.Sprintf("players/%v/rests/end", playerID), request).Post()

	return
}
