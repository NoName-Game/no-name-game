package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	"bitbucket.org/no-name-game/nn-grpc/build/pb"
)

// =================
// Current Controller Cache
// =================

// GetCurrentControllerCache - Metodo generico per il recupero degli stati di un player
func GetCurrentControllerCache(playerID uint32) (result string, err error) {
	if result, err = config.App.Redis.Connection.Get(fmt.Sprintf("player_%v_current_controller", playerID)).Result(); err != nil {
		err = errors.New("cached state not found")
	}

	return
}

// SetCurrentControllerCache - Metodo generico per il settaggio di uno stato in memoria di un determinato player
func SetCurrentControllerCache(playerID uint32, controller string) {
	if err := config.App.Redis.Connection.Set(fmt.Sprintf("player_%v_current_controller", playerID), controller, 0).Err(); err != nil {
		panic(err)
	}
}

// DelCurrentControllerCache - Metodo generico per la cancellazione degli stati di un determinato player
func DelCurrentControllerCache(playerID uint32) {
	if err := config.App.Redis.Connection.Del(fmt.Sprintf("player_%v_current_controller", playerID)).Err(); err != nil {
		panic(err)
	}
}

// =================
// Cache - Controller
// =================
type ControllerCacheData struct {
	Stage   int32
	Payload string
}

func SetControllerCacheData(playerID uint32, controller string, stage int32, payload interface{}) (err error) {
	var marshalPayload []byte
	if marshalPayload, err = json.Marshal(payload); err != nil {
		return fmt.Errorf("error marshal controller payload data for cache: %s", err.Error())
	}

	var marshalData []byte
	if marshalData, err = json.Marshal(ControllerCacheData{
		Stage:   stage,
		Payload: string(marshalPayload),
	}); err != nil {
		return fmt.Errorf("error marshal data for cache: %s", err.Error())
	}

	return config.App.Redis.Connection.Set(fmt.Sprintf("player_%v_controller_%s", playerID, controller), marshalData, 72*time.Hour).Err()
}

func GetControllerCacheData(playerID uint32, controller string, payload interface{}) (stage int32, err error) {
	var result string
	result, _ = config.App.Redis.Connection.Get(fmt.Sprintf("player_%v_controller_%s", playerID, controller)).Result()

	var cachedData ControllerCacheData
	if result != "" {
		if err = json.Unmarshal([]byte(result), &cachedData); err == nil {
			stage = cachedData.Stage
			err = json.Unmarshal([]byte(cachedData.Payload), &payload)
		}
	}

	return
}

func DelControllerCacheData(playerID uint32, controller string) (err error) {
	if err := config.App.Redis.Connection.Del(fmt.Sprintf("player_%v_controller_%s", playerID, controller)).Err(); err != nil {
		return fmt.Errorf("cant delete controller cache data: %s", err.Error())
	}
	return
}

// =================
// Cache Map
// =================

// SetMapInCache - Salvo mappa in cache per non appesantire le chiamate a DB
func SetMapInCache(maps *pb.Maps) (err error) {
	var jsonValue []byte
	jsonValue, _ = json.Marshal(maps)

	if err := config.App.Redis.Connection.Set(fmt.Sprintf("hunting_map_%v", maps.ID), string(jsonValue), 0).Err(); err != nil {
		return fmt.Errorf("cant set map in cache: %s", err.Error())
	}
	return
}

// GetMapInCache - Recupera mappa in memoria
func GetMapInCache(MapID uint32) (maps *pb.Maps, err error) {
	var result string
	if result, err = config.App.Redis.Connection.Get(fmt.Sprintf("hunting_map_%v", MapID)).Result(); err != nil {
		return maps, fmt.Errorf("cant get map in cache: %s", err.Error())
	}

	err = json.Unmarshal([]byte(result), &maps)
	return
}
