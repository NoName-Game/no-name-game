package helpers

import (
	"encoding/json"

	"nn-grpc/build/pb"
)

// DecodeMapToDisplay - Converte la logica della mappa in qualcosa di visibilie
// per il client mostrando con diversi caratteri le superfici, mob e player
func DecodeMapToDisplay(maps *pb.PlanetMap, playerPositionX int32, playerPositionY int32) (result string, err error) {
	// Setto cornice di apertura
	result = "<code>+---------------------+\n"

	// Recupero mappa
	var cellGrid [][]bool
	if err = json.Unmarshal([]byte(maps.CellGrid), &cellGrid); err != nil {
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
func CheckForMob(maps *pb.PlanetMap, x int32, y int32) (enemy *pb.Enemy, result bool) {
	for i := 0; i < len(maps.Enemies); i++ {
		if x == maps.Enemies[i].PositionX && y == maps.Enemies[i].PositionY {
			return maps.Enemies[i], true
		}
	}

	return
}

// CheckForTresure - Verifica posizione dei tesori
func CheckForTresure(maps *pb.PlanetMap, x int32, y int32) (tresure *pb.Tresure, result bool) {
	for i := 0; i < len(maps.Tresures); i++ {
		if x == maps.Tresures[i].PositionX && y == maps.Tresures[i].PositionY {
			return maps.Tresures[i], true
		}
	}

	return
}
