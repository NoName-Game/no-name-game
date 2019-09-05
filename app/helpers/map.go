package helpers

import (
	"encoding/json"

	"bitbucket.org/no-name-game/no-name/services"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
)

func TextDisplay(m nnsdk.Map) string {
	result := "+---------------------+\n"
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
					} else if x == m.EnemyX && y == m.EnemyY {
						result += "*"
					} else {
						result += " "
					}

				}
			} else {
				//log.Println("Out of Bound X: ", x, " Y: ", y)
				result += "#"
			}

		}
		result += "|"
		result += "\n"
	}
	result += "+---------------------+"
	return result
}
