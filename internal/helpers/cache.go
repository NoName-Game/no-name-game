package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/config"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
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
	marshalPayload, _ := json.Marshal(payload)

	marshalData, _ := json.Marshal(ControllerCacheData{
		Stage:   stage,
		Payload: string(marshalPayload),
	})

	return config.App.Redis.Connection.Set(fmt.Sprintf("player_%v_controller_%s", playerID, controller), marshalData, 60*time.Minute).Err()
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

func DelControllerCacheData(playerID uint32, controller string) {
	if err := config.App.Redis.Connection.Del(fmt.Sprintf("player_%v_controller_%s", playerID, controller)).Err(); err != nil {
		panic(err)
	}
}

// =================
// Cache Map
// =================

// SetMapInCache - Salvo mappa in cache per non appesantire le chiamate a DB
func SetMapInCache(maps *pb.Maps) {
	var jsonValue []byte
	jsonValue, _ = json.Marshal(maps)

	if err := config.App.Redis.Connection.Set(fmt.Sprintf("hunting_map_%v", maps.ID), string(jsonValue), 0).Err(); err != nil {
		panic(err)
	}
}

// GetMapInCache - Recupera mappa in memoria
func GetMapInCache(MapID uint32) (maps *pb.Maps, err error) {
	var result string
	if result, err = config.App.Redis.Connection.Get(fmt.Sprintf("hunting_map_%v", MapID)).Result(); err != nil {
		err = errors.New("cached state not found")
	}

	err = json.Unmarshal([]byte(result), &maps)
	return
}
