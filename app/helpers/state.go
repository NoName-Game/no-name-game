package helpers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
	"bitbucket.org/no-name-game/nn-telegram/app/providers"

	"bitbucket.org/no-name-game/nn-telegram/services"
)

// GetRedisState - set function state in Redis
func GetRedisState(player nnsdk.Player) string {
	var route string
	route, _ = services.Redis.Get(strconv.FormatUint(uint64(player.ID), 10)).Result()

	return route
}

// SetRedisState - set function state in Redis
func SetRedisState(player nnsdk.Player, function string) {
	err := services.Redis.Set(strconv.FormatUint(uint64(player.ID), 10), function, 0).Err()
	if err != nil {
		services.ErrorHandler("Error SET player state in redis", err)
	}
}

// GetHuntingRedisState - get hunting state in Redis
func GetHuntingRedisState(IDMap uint, player nnsdk.Player) (huntingMap nnsdk.Map) {
	state, err := services.Redis.Get(fmt.Sprintf("hunting_%v_%v", IDMap, player.ID)).Result()
	if err != nil {
		services.ErrorHandler("Error getting hunting state in redis", err)
	}

	json.Unmarshal([]byte(state), &huntingMap)
	return
}

// SetRedisState - set function state in Redis
func SetHuntingRedisState(IDMap uint, player nnsdk.Player, value interface{}) {
	jsonValue, _ := json.Marshal(value)
	err := services.Redis.Set(fmt.Sprintf("hunting_%v_%v", IDMap, player.ID), string(jsonValue), 0).Err()
	if err != nil {
		services.ErrorHandler("Error SET player state in redis", err)
	}
}

// DelRedisState - del function state in Redis
func DelRedisState(player nnsdk.Player) {
	err := services.Redis.Del(strconv.FormatUint(uint64(player.ID), 10)).Err()
	if err != nil {
		services.ErrorHandler("Error DEL player state in redis", err)
	}
}

// StartAndCreatePlayerState - create and set redis state
func StartAndCreatePlayerState(route string, player nnsdk.Player) (playerState nnsdk.PlayerState) {
	playerState, _ = GetPlayerStateByFunction(player, route)

	if playerState.ID < 1 {
		newPlayerState := nnsdk.PlayerState{
			Function: route,
			PlayerID: player.ID,
		}

		playerState, _ = providers.CreatePlayerState(newPlayerState)
	}

	SetRedisState(player, route)
	return
}

// CheckState - create and set redis state
func CheckState(route string, payload interface{}, player nnsdk.Player) (playerState nnsdk.PlayerState, isNewState bool) {
	playerState, _ = GetPlayerStateByFunction(player, route)

	if playerState.ID < 1 {
		jsonPayload, _ := json.Marshal(payload)
		newPlayerState := nnsdk.PlayerState{
			Function: route,
			PlayerID: player.ID,
			Payload:  string(jsonPayload),
		}

		playerState, _ = providers.CreatePlayerState(newPlayerState)
		isNewState = true
	}

	SetRedisState(player, route)
	return
}

// FinishAndCompleteState - finish and set completed in playerstate
func FinishAndCompleteState(state nnsdk.PlayerState, player nnsdk.Player) {
	// Stupid poninter stupid json pff
	t := new(bool)
	*t = true

	state.Completed = t
	state, _ = providers.UpdatePlayerState(state) // Update
	state, _ = providers.DeletePlayerState(state) // Delete

	DelRedisState(player)
}

// DeleteRedisAndDbState - delete redis and db state
func DeleteRedisAndDbState(player nnsdk.Player) {
	rediState := GetRedisState(player)

	if rediState != "" {
		playerState, _ := GetPlayerStateByFunction(player, rediState)
		_, err := providers.DeletePlayerState(playerState) // Delete
		if err != nil {
			services.ErrorHandler("Error delete player state", err)
		}
	}

	DelRedisState(player)
}

// UnmarshalPayload - Unmarshal payload state
func UnmarshalPayload(payload string, funcInterface interface{}) {
	if payload != "" {
		err := json.Unmarshal([]byte(payload), &funcInterface)
		if err != nil {
			services.ErrorHandler("Error unmarshal payload", err)
		}
	}
}
