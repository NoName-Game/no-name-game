package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"

	"bitbucket.org/no-name-game/nn-telegram/services"
)

// =================
// Cache state
// =================

// GetCacheState - Metodo generico per il recupero degli stati di un player
func GetCacheState(playerID uint32) (result string, err error) {
	if result, err = services.Redis.Get(fmt.Sprintf("current_state_player_%v", playerID)).Result(); err != nil {
		err = errors.New("cached state not found")
	}

	return
}

// SetCacheState - Metodo generico per il settaggio di uno stato in memoria di un determinato player
func SetCacheState(playerID uint32, controller string) {
	if err := services.Redis.Set(fmt.Sprintf("current_state_player_%v", playerID), controller, 0).Err(); err != nil {
		panic(err)
	}
}

// DelCacheState - Metodo generico per la cancellazione degli stati di un determinato player
func DelCacheState(playerID uint32) {
	if err := services.Redis.Del(fmt.Sprintf("current_state_player_%v", playerID)).Err(); err != nil {
		panic(err)
	}
}

// =================
// Cache Controller
// =================

func GetCacheControllerStage(playerID uint32, controller string) (result string, err error) {
	if result, err = services.Redis.Get(fmt.Sprintf("player_%v_controller_%s", playerID, controller)).Result(); err != nil {
		err = errors.New("cached state not found")
	}

	return
}

func SetCacheControllerStage(playerID uint32, controller string, stage int32) {
	if err := services.Redis.Set(fmt.Sprintf("player_%v_controller_%s", playerID, controller), stage, 0).Err(); err != nil {
		panic(err)
	}
}

func DelCacheControllerStage(playerID uint32, controller string) {
	if err := services.Redis.Del(fmt.Sprintf("player_%v_controller_%s", playerID, controller)).Err(); err != nil {
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

	if err := services.Redis.Set(fmt.Sprintf("hunting_map_%v", maps.ID), string(jsonValue), 0).Err(); err != nil {
		panic(err)
	}
}

// GetMapInCache - Recupera mappa in memoria
func GetMapInCache(MapID uint32) (maps *pb.Maps, err error) {
	var result string
	if result, err = services.Redis.Get(fmt.Sprintf("hunting_map_%v", MapID)).Result(); err != nil {
		err = errors.New("cached state not found")
	}

	err = json.Unmarshal([]byte(result), &maps)
	return
}

// =================
// Cache - Payload
// =================
func SetPayloadController(playerID uint32, controller string, data interface{}) {
	marshalData, _ := json.Marshal(data)

	if err := services.Redis.Set(fmt.Sprintf("player_%v_controller_%s_payload", playerID, controller), marshalData, 60*time.Minute).Err(); err != nil {
		panic(err)
	}
}

func GetPayloadController(playerID uint32, controller string, payload interface{}) (err error) {
	var result string
	result, _ = services.Redis.Get(fmt.Sprintf("player_%v_controller_%s_payload", playerID, controller)).Result()

	if result != "" {
		return json.Unmarshal([]byte(result), &payload)
	}

	return
}

func DelPayloadController(playerID uint32, controller string) {
	if err := services.Redis.Del(fmt.Sprintf("player_%v_controller_%s_payload", playerID, controller)).Err(); err != nil {
		panic(err)
	}
}
