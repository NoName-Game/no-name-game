package helpers

import (
	"encoding/json"

	"bitbucket.org/no-name-game/no-name/services"

	"bitbucket.org/no-name-game/no-name/app/acme/nnsdk"
)

func TextDisplay(m nnsdk.Map) string {
	result := "┏━━━━━━━━━━━━━━━━━━┓\n"
	var cellMap [66][66]bool
	err := json.Unmarshal([]byte(m.Cell), &cellMap)
	if err != nil {
		services.ErrorHandler("UnMarshal Error", err)
	}
	for x := 15; x < 26; x++ {
		result += "┃"
		for y := 15; y < 36; y++ {
			if cellMap[x][y] {
				result += "#"
			} else {
				result += " "
			}
		}
		result += "┃"
		result += "\n"
	}
	result += "┗━━━━━━━━━━━━━━━━━━┛"
	return result
}
