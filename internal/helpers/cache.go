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

	_, err = config.App.Server.Connection.CreateTelegramStatus(NewContext(1), &pb.CreateTelegramStatusRequest{
		PlayerID:   playerID,
		Stage:      stage,
		Controller: controller,
		Payload:    string(marshalPayload),
	})

	return
}

func GetControllerCacheData(playerID uint32, controller string, payload interface{}) (stage int32, err error) {
	var rGetTelegramStatus *pb.GetTelegramStatusResponse
	if rGetTelegramStatus, err = config.App.Server.Connection.GetTelegramStatus(NewContext(1), &pb.GetTelegramStatusRequest{
		PlayerID:   playerID,
		Controller: controller,
	}); err != nil {
		return 0, nil
	}

	if err = json.Unmarshal([]byte(rGetTelegramStatus.GetPayload()), &payload); err != nil {
		return
	}

	return rGetTelegramStatus.GetStage(), nil

}

func DelControllerCacheData(playerID uint32, controller string) (err error) {
	if _, err = config.App.Server.Connection.DeleteTelegramStatus(NewContext(1), &pb.DeleteTelegramStatusRequest{
		PlayerID:   playerID,
		Controller: controller,
	}); err != nil {
		return fmt.Errorf("cant delete controller cache data: %s", err.Error())
	}

	return nil
}

// =================
// Cache Map
// =================
// SetMapInCache - Salvo mappa in cache per non appesantire le chiamate a DB
func SetMapInCache(maps *pb.PlanetMap) (err error) {
	var jsonValue []byte
	jsonValue, _ = json.Marshal(maps)

	if err := config.App.Redis.Connection.Set(fmt.Sprintf("hunting_map_%v", maps.ID), string(jsonValue), 0).Err(); err != nil {
		return fmt.Errorf("cant set map in cache: %s", err.Error())
	}
	return
}

// GetMapInCache - Recupera mappa in memoria
func GetMapInCache(MapID uint32) (planetMap *pb.PlanetMap, err error) {
	var result string
	if result, err = config.App.Redis.Connection.Get(fmt.Sprintf("hunting_map_%v", MapID)).Result(); err != nil {
		return planetMap, fmt.Errorf("cant get map in cache: %s", err.Error())
	}

	err = json.Unmarshal([]byte(result), &planetMap)
	return
}

// =================
// Cache Player Position
// =================
// SetPlayerPlanetPositionInCache - Salvo posizione player in cache per non appesantire le chiamate a DB
func SetPlayerPlanetPositionInCache(playerID uint32, planet *pb.Planet) (err error) {
	var jsonValue []byte
	jsonValue, _ = json.Marshal(planet)

	if err := config.App.Redis.Connection.Set(fmt.Sprintf("player_%v_current_planet", playerID), string(jsonValue), 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("cant set player position in cache: %s", err.Error())
	}
	return
}

// GetPlayerPlanetPositionInCache - Recupera mappa in memoria
func GetPlayerPlanetPositionInCache(playerID uint32) (planet *pb.Planet, err error) {
	var result string
	if result, err = config.App.Redis.Connection.Get(fmt.Sprintf("player_%v_current_planet", playerID)).Result(); err != nil {
		return planet, fmt.Errorf("cant get player position in cache: %s", err.Error())
	}

	err = json.Unmarshal([]byte(result), &planet)
	return
}

func DelPlayerPlanetPositionInCache(playerID uint32) (err error) {
	if err := config.App.Redis.Connection.Del(fmt.Sprintf("player_%v_current_planet", playerID)).Err(); err != nil {
		return fmt.Errorf("cant delete player position in cache data: %s", err.Error())
	}
	return
}

// =================
// Exploration Categories
// =================
// SetExplorationCategoriesInCache - Setto categorie esplorazione
func SetExplorationCategoriesInCache(categories []*pb.ExplorationCategory) (err error) {
	var jsonValue []byte
	jsonValue, _ = json.Marshal(categories)

	if err := config.App.Redis.Connection.Set("sys_exploration_categories", string(jsonValue), 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("cant set exploration categories in cache: %s", err.Error())
	}
	return
}

// GetExplorationCategoriesInCache - Recupero categorie esplorazione
func GetExplorationCategoriesInCache() (categories []*pb.ExplorationCategory, err error) {
	var result string
	if result, err = config.App.Redis.Connection.Get("sys_exploration_categories").Result(); err != nil {
		return categories, fmt.Errorf("cant get exploration categories in cache: %s", err.Error())
	}

	err = json.Unmarshal([]byte(result), &categories)
	return
}

// =================
// AntiFlood
// =================
// SetAntiFlood
func SetAntiFlood(playerID uint32) (err error) {
	if err := config.App.Redis.Connection.Set(fmt.Sprintf("player_%v_antiflood", playerID), 1, 5*time.Second).Err(); err != nil {
		return fmt.Errorf("cant set antiflood in cache: %s", err.Error())
	}
	return
}

// GetAntiFlood
func GetAntiFlood(playerID uint32) (result string, err error) {
	if result, err = config.App.Redis.Connection.Get(fmt.Sprintf("player_%v_antiflood", playerID)).Result(); err != nil {
		return result, fmt.Errorf("cant get player antiflood in cache: %s", err.Error())
	}

	return
}

// DelAntiFlood
func DelAntiFlood(playerID uint32) (err error) {
	if err := config.App.Redis.Connection.Del(fmt.Sprintf("player_%v_antiflood", playerID)).Err(); err != nil {
		return fmt.Errorf("cant delete player antiflood in cache : %s", err.Error())
	}
	return
}
