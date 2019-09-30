package helpers

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
	"bitbucket.org/no-name-game/no-name/app/providers"

	"bitbucket.org/no-name-game/no-name/services"
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

// DelRedisState - del function state in Redis
func DelRedisState(player nnsdk.Player) {
	err := services.Redis.Del(strconv.FormatUint(uint64(player.ID), 10)).Err()
	if err != nil {
		services.ErrorHandler("Error DEL player state in redis", err)
	}
}

// StartAndCreatePlayerState - create and set redis state
func StartAndCreatePlayerState(route string, player nnsdk.Player) (playerState nnsdk.PlayerState) {
	playerState = GetPlayerStateByFunction(player, route)

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
		playerState := GetPlayerStateByFunction(player, rediState)
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
