package godeffect

import (
	"log"

	"bitbucket.org/no-name-game/no-name/app/commands/godeffect/galaxies"
	"bitbucket.org/no-name-game/no-name/app/models"
)

// OMG - Simulate user registration
func OMG() {
	nUsers := 100
	for deviation := 1; deviation <= nUsers; deviation++ {
		size := 10.0
		sphere := galaxies.Sphere{
			Size:      size,
			Density:   20,
			Deviation: float64(deviation),
		}

		for _, star := range sphere.Generate() {

			log.Println(star)

			newStar := models.Star{
				Name:        star.Name,
				Size:        star.Size,
				X:           star.Position[0],
				Y:           star.Position[1],
				Z:           star.Position[2],
				Temperature: star.Temperature,
				Color:       star.Color,
			}

			newStar.Create()
		}
	}
}
