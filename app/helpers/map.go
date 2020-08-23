package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/no-name-game/nn-telegram/services"
	gocache "github.com/patrickmn/go-cache"

	pb "bitbucket.org/no-name-game/nn-grpc/build/proto"
)

// DecodeMapToDisplay - Converte la logica della mappa in qualcosa di visibilie
// per il client mostrando con diversi caratteri le superfici, mob e player
func DecodeMapToDisplay(maps *pb.Maps, playerPositionX int32, playerPositionY int32) (result string, err error) {
	// Setto cornice di apertura
	result = "<code>+---------------------+\n"

	// Recupero mappa
	var cellGrid [][]bool
	err = json.Unmarshal([]byte(maps.CellGrid), &cellGrid)
	if err != nil {
		return result, err
	}

	// Ciclo tutta la griglia andando a sostituire tutte le occorrenze
	for x := playerPositionX - 5; x < playerPositionX+5; x++ { // 11
		// Inserisco cornice della riga
		result += "|"
		// Conto quanto è lunga la mappa
		mapWith := int32(len(cellGrid[0]))
		for y := playerPositionY - 10; y < playerPositionY+11; y++ { // 21
			// Conto quanto è alta la mappa
			mapHeight := int32(len(cellGrid[0]))
			// Verifico che siamo all'interno dei limiti
			if (x >= 0 && x < mapWith) && (y >= 0 && y < mapHeight) { // In bounds
				// Se è true come da regola vuol dire che NON è calpestabile
				if cellGrid[x][y] {
					// Lo gestisco come terreno NON calpestabile
					result += "#"
				} else {
					// Se corrisponde alla posizione del Player lo mostro
					if x == playerPositionX && y == playerPositionY {
						result += "P"
						continue
					}

					// +---------------------+

					// Renderizzo mob sulla mppa
					_, isMob := CheckForMob(maps, x, y)
					if isMob {
						result += "M"
						continue
					}

					// Renderizzo mob sulla mppa
					_, isTresure := CheckForTresure(maps, x, y)
					if isTresure {
						result += "T"
						continue
					}

					result += " "
				}
			} else {
				result += "." // Delimito i bordi
			}
		}
		result += "|"
		result += "\n"
	}

	// Cornice di chiusra
	result += "+---------------------+</code>"

	return
}

// CheckForMob - Verifica posizione dei mob
func CheckForMob(maps *pb.Maps, x int32, y int32) (enemy *pb.Enemy, result bool) {
	for i := 0; i < len(maps.Enemies); i++ {
		if x == maps.Enemies[i].PositionX && y == maps.Enemies[i].PositionY {
			return maps.Enemies[i], true
		}
	}

	return
}

// CheckForTresure - Verifica posizione dei tesori
func CheckForTresure(maps *pb.Maps, x int32, y int32) (tresure *pb.Tresure, result bool) {
	for i := 0; i < len(maps.Tresures); i++ {
		if x == maps.Tresures[i].PositionX && y == maps.Tresures[i].PositionY {
			return maps.Tresures[i], true
		}
	}

	return
}

// ChooseMob - viene richiamato principalmente dalla mappa, la sua funzione è
// quella di ritornare un mob dalla mappa tra quelli vicini al player
func ChooseEnemyInMap(maps *pb.Maps, playerPositionX int32, playerPositionY int32) (enemyID int32, err error) {
	// Recupero mappa
	var cellGrid [][]bool
	err = json.Unmarshal([]byte(maps.CellGrid), &cellGrid)
	if err != nil {
		return enemyID, err
	}

	for x := playerPositionX - 5; x < playerPositionX+5; x++ {
		// Conto quanto è lunga la mappa
		mapWith := int32(len(cellGrid[0]))
		for y := playerPositionY - 10; y < playerPositionY+11; y++ {
			// Conto quanto è alta la mappa
			mapHeight := int32(len(cellGrid[0]))
			if (x >= 0 && x < mapWith) && (y >= 0 && y < mapHeight) { // In bounds
				for i := 0; i < len(maps.Enemies); i++ {
					if x == maps.Enemies[i].PositionX && y == maps.Enemies[i].PositionY {
						return int32(i), nil
					}
				}

			}
		}
	}
	return -1, err
}

// SetMapInCache - Salvo mappa in cache per non appesantire le chiamate a DB
func SetMapInCache(maps *pb.Maps) {
	var jsonValue []byte
	jsonValue, _ = json.Marshal(maps)

	services.Cache.Set(
		fmt.Sprintf("hunting_map_%v", maps.ID),
		string(jsonValue),
		gocache.NoExpiration,
	)
}

// GetMapInCache - Recupera mappa in memoria
func GetMapInCache(MapID uint32) (maps *pb.Maps, err error) {
	record, found := services.Cache.Get(fmt.Sprintf("hunting_map_%v", MapID))
	if found {
		err = json.Unmarshal([]byte(record.(string)), &maps)
		return
	}

	err = errors.New("map in cache not found")
	return
}

// GetCachedPlayerPositionInMap - recupero posizione di una player in una specifica mappa
func GetCachedPlayerPositionInMap(maps *pb.Maps, player *pb.Player, positionType string) (value int32, err error) {
	record, found := services.Cache.Get(fmt.Sprintf(
		"hunting_map_%v_player_%v_position_%s",
		maps.ID,
		player.ID,
		positionType,
	))

	if found {
		return record.(int32), nil
	}

	err = errors.New("cached state not found")
	return
}

// SetCachedPlayerPositionInMap - Imposto posizione di un player su una determinata mappa
func SetCachedPlayerPositionInMap(maps *pb.Maps, player *pb.Player, positionType string, value int32) {
	services.Cache.Set(
		fmt.Sprintf("hunting_map_%v_player_%v_position_%s", maps.ID, player.ID, positionType),
		value,
		60*time.Minute,
	)
}
