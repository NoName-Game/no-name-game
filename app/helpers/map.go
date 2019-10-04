package helpers

import (
	"encoding/json"

	"bitbucket.org/no-name-game/nn-telegram/services"

	"bitbucket.org/no-name-game/nn-telegram/app/acme/nnsdk"
)

func TextDisplay(m nnsdk.Map) string {
	result := "<code>+---------------------+\n"
	var cellMap [66][66]bool
	err := json.Unmarshal([]byte(m.Cell), &cellMap)
	if err != nil {
		services.ErrorHandler("UnMarshal Error", err)
	}
	//log.Println("Player X: ", m.PlayerX, " Y: ", m.PlayerY)
	for x := m.PlayerX - 5; x < m.PlayerX+5; x++ { //11
		result += "|"
		for y := m.PlayerY - 10; y < m.PlayerY+11; y++ { // 21
			if (x >= 0 && x < 66) && (y >= 0 && y < 66) { // In bounds
				if cellMap[x][y] {
					result += "#"
				} else {
					if x == m.PlayerX && y == m.PlayerY {
						result += "@"
						continue
					} else if checkForMob(m, x, y) {
						result += "*"
					} else {
						result += " "
					}
				}
			} else {
				result += "#"
			}

		}
		result += "|"
		result += "\n"
	}
	result += "+---------------------+</code>"
	return result
}

func checkForMob(m nnsdk.Map, x, y int) bool {
	for i := 0; i < len(m.Enemies); i++ {
		if x == m.Enemies[i].MapPositionX && y == m.Enemies[i].MapPositionY {
			return true
		}
	}
	return false
}

func ChooseMob(m nnsdk.Map) int {
	for x := m.PlayerX - 5; x < m.PlayerX+5; x++ {
		for y := m.PlayerY - 10; y < m.PlayerY+11; y++ {
			if (x >= 0 && x < 66) && (y >= 0 && y < 66) { // In bounds
				for i := 0; i < len(m.Enemies); i++ {
					if x == m.Enemies[i].MapPositionX && y == m.Enemies[i].MapPositionY {
						return i
					}
				}

			}
		}
	}
	return -1
}
